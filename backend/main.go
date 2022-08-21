package main

import (
	"backend-reg/api"
	"backend-reg/mydb"
)

func main() {

	db := mydb.Conn()
	defer db.Close()

	mydb.Automigrate(db)

	api.RegisterRoutes()

}
