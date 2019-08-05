package database

import (
	// "context"
	"database/sql"
	"fmt"
	// "strings"

	_ "github.com/lib/pq"

	"github.com/sjsafranek/ligneous"
)

var (
	logger = ligneous.AddLogger("database", "trace", "./log/find5")
)

func New(connString string) *Database {
	return &Database{connString: connString}
}

type Database struct {
	connString string
}

func (self *Database) GetConnection() (*sql.DB, error) {
	// TODO
	//  - use connection pool
	return sql.Open("postgres", self.connString)
}

func (self *Database) Exec(clbk func(*sql.DB) error) error {
	// TODO
	//  - use connection pool
	db, err := sql.Open("postgres", self.connString)
	defer db.Close()
	if nil != err {
		return err
	}
	return clbk(db)
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

	db, err := self.GetConnection()
	if nil != err {
		return &User{}, err
	}

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
		return &user, err
	}

	c := 0
	for rows.Next() {
		var temp string
		rows.Scan(&temp)
		user.Unmarshal(temp)
		c++
	}

	if 0 == c {
		return &user, fmt.Errorf("Not found")
	}

	user.db = self
	return &user, err
}

func (self *Database) Insert(query string, args ...interface{}) error {

	db, err := self.GetConnection()
	if nil != err {
		return err
	}
	defer db.Close()

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
}
