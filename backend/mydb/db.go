package mydb

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func Conn() *sql.DB {

	db, err := sql.Open("mysql", "post_api:post_api@tcp(127.0.0.1:6603)/post_api")

	if err != nil {
		log.Fatal(err)
	}

	return db
}

func Automigrate(db *sql.DB) {

	const tableCreationQuery = `CREATE TABLE IF NOT EXISTS users
		(
			id SERIAL,
			username CHAR(255),
			email CHAR(255),
			password CHAR(255)
		)`
	_ = db.QueryRow(tableCreationQuery)

	// error handling #TODO

}
