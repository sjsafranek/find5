package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"github.com/sjsafranek/logger"
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

func (self *Database) GetVersion() (string, error) {
	var version string
	return version, self.Exec(func(db *sql.DB) error {
		rows, err := db.Query(`SELECT value FROM config WHERE key='version'`)
		if nil != err {
			return err
		}

		for rows.Next() {
			rows.Scan(&version)
		}

		return nil
	})
}

func (self *Database) Exec(clbk func(*sql.DB) error) error {
	return clbk(self.db)
}

func (self *Database) QueryRow(query string, args ...interface{}) *sql.Row {
	logger.Debug(query, args)
	return self.db.QueryRow(query, args...)
}

func (self *Database) Insert(query string, args ...interface{}) error {
	logger.Debug(query, args)
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

/**
 * APP FUNCTIONS
 */
func (self *Database) getUser(query string, args ...interface{}) (*User, error) {
	var user User

	logger.Debug(query, args)

	var temp string
	err := self.db.QueryRow(query, args...).Scan(&temp)
	if nil != err {
		if "sql: no rows in result set" == err.Error() {
			return &user, errors.New("Not found")
		}
		return &user, err
	}

	err = user.Unmarshal(temp)
	if nil != err {
		return &user, err
	}

	user.db = self
	return &user, nil
}

func (self *Database) CreateUser(email, username string) (*User, error) {
	query := `
		INSERT INTO users (email, username) VALUES ($1, $2) RETURNING json_build_object(
			'email', email,
			'username', username,
			'apikey', apikey,
			'secret_token', secret_token,
			'is_active', is_active,
			'is_deleted', is_deleted,
			'created_at', to_char(created_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"'),
			'updated_at', to_char(updated_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"')
		);
	`
	return self.getUser(query, email, username)
}

func (self *Database) CreateUserIfNotExists(email, username string) (*User, error) {
	user, err := self.CreateUser(email, username)
	if nil != err && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return self.GetUserByUsername(username)
	}
	return user, nil
}

//
func (self *Database) getUserBy(key, value string) (*User, error) {
	query := fmt.Sprintf(`
		SELECT
			user_json
		FROM users_view
		WHERE %v = $1;
	`, key)
	return self.getUser(query, value)
}

func (self *Database) GetUserByEmail(email string) (*User, error) {
	return self.getUserBy("email", email)
}

func (self *Database) GetUserByUsername(username string) (*User, error) {
	return self.getUserBy("username", username)
}

func (self *Database) GetUserByApikey(apikey string) (*User, error) {
	return self.getUserBy("apikey", apikey)
}

func (self *Database) GetUsers() ([]*User, error) {
	var users []*User
	return users, self.Exec(func(db *sql.DB) error {

		rows, err := db.Query(`
			SELECT
				json_agg(user_json)
			FROM users_view
			WHERE
				is_deleted = false
			;`)
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
