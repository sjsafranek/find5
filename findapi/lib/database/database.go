package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/sjsafranek/ligneous"
)

var (
	logger = ligneous.AddLogger("database", "trace", "./log/find5")
)

func New(connString string) *Database {
	db, err := sql.Open("postgres", connString)
	if nil != err {
		panic(err)
	}
	// db.SetMaxOpenConns(6) // <- default is unlimited)
	db.SetMaxIdleConns(6)
	// db.SetConnMaxLifetime(2 * time.Minute)
	return &Database{connString: connString, db: db}
}

type Database struct {
	connString string
	db         *sql.DB // <- built in connection pool
}

func (self *Database) Exec(clbk func(*sql.DB) error) error {
	return clbk(self.db)
}

func (self *Database) CreateUser(email, username, password string) (*User, error) {
	query := `INSERT INTO users (email, username, password) VALUES ($1, $2, $3);`
	err := self.Insert(query, email, username, password)
	if nil != err {
		return &User{}, err
	}

	return self.GetUserFromUsername(username)
}

func (self *Database) GetUserFromUsername(username string) (*User, error) {
	return self.getUser("username", username)
}

func (self *Database) GetUserFromApikey(apikey string) (*User, error) {
	return self.getUser("apikey", apikey)
}

func (self *Database) getUser(column string, value string) (*User, error) {
	var user User
	return &user, self.Exec(func(db *sql.DB) error {
		query := fmt.Sprintf(`
				SELECT row_to_json(u)
					FROM (
					    SELECT
					        email,
					        username,
							apikey,
					        secret_token,
					        is_deleted,
					        to_char(created_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"') as created_at,
					        to_char(updated_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"') as updated_at
					    FROM users
					    WHERE %v=$1
						AND is_deleted = false
					) AS u ;`, column)

		rows, err := db.Query(query, value)
		if nil != err {
			return err
		}

		c := 0
		for rows.Next() {
			var temp string
			rows.Scan(&temp)
			user.Unmarshal(temp)
			c++
		}

		if 0 == c {
			return errors.New("Not found")
		}

		user.db = self
		return nil
	})
}

func (self *Database) GetUsers() ([]*User, error) {
	var users []*User
	return users, self.Exec(func(db *sql.DB) error {
		rows, err := db.Query(`
			SELECT json_agg(c) FROM (
			    SELECT
			        email,
			        username,
			        apikey,
			        secret_token,
			        is_deleted,
			        to_char(created_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"') as created_at,
			        to_char(updated_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"') as updated_at
			    FROM users
			    WHERE is_deleted = false
			) c;`)
		if nil != err {
			return err
		}

		for rows.Next() {
			var temp string
			rows.Scan(&temp)
			err = json.Unmarshal([]byte(temp), &users)
			if nil != err {
				return err
			}
		}

		for i := range users {
			users[i].db = self
		}

		return nil
	})
}

func (self *Database) Insert(query string, args ...interface{}) error {
	return self.Exec(func(db *sql.DB) error {
		tx, err := db.Begin()
		if err != nil {
			tx.Rollback()
			logger.Error(err)
			return err
		}

		stmt, err := tx.Prepare(query)
		if err != nil {
			tx.Rollback()
			logger.Error(err)
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(args...)
		if nil != err {
			tx.Rollback()
			logger.Error(err)
			return err
		}

		tx.Commit()
		return nil
	})
}
