package main

import (
	"final_sprint/config"
	"final_sprint/database"
	"final_sprint/server"
	"log"
	"os"
)

func main() {
	port := config.GetPort()
	dbFile := config.GetDBFile()

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		log.Printf("Создание базы данных по пути: %s", dbFile)
		if err := database.CreateDatabase(dbFile); err != nil {
			log.Fatalf("Ошибка при создании базы данных: %v", err)
		}
		log.Println("База данных и таблица успешно созданы")
	} else {
		log.Println("База данных уже существует")
	}

	server.StartServer(port)
}
