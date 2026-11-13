package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/Miroku77/goprxveil/internal/api"
	"github.com/Miroku77/goprxveil/internal/browser"
	"github.com/Miroku77/goprxveil/internal/config"
	"github.com/Miroku77/goprxveil/internal/handlers"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println(err)
	}
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	handlers.InitDB(db)
	go api.StartServer()
	go browser.JobCallShop(db)
	select {}
}
