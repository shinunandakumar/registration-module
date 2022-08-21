package mydb

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func Conn() *sql.DB {

	db, err := sql.Open("mysql", "post_api:post_api@tcp(db:3306)/post_api")

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
			name CHAR(255) NULL,
			email CHAR(255) NULL,
			phone CHAR(255) NULL,
			password CHAR(255)
		)`
	_ = db.QueryRow(tableCreationQuery)

	// error handling #TODO

}
