package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"final_sprint/api"
	"final_sprint/config"
	"final_sprint/database"
	"final_sprint/server"
)

func getToken(password string) (string, error) {
	creds := map[string]string{"password": password}
	body, err := json.Marshal(creds)
	if err != nil {
		return "", err
	}

	resp, err := http.Post("http://localhost:7540/api/signin", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get token, status code: %d", resp.StatusCode)
	}

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	token, ok := result["token"]
	if !ok {
		return "", fmt.Errorf("token not found in response")
	}

	return token, nil
}

func main() {
	config.InitConfig()

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

	db, err := database.InitDB(dbFile)
	if err != nil {
		log.Fatalf("Ошибка инициализации базы данных: %v", err)
	}
	defer db.Close()

	taskHandler := api.NewTaskHandler(db)

	http.HandleFunc("/add", server.Auth(taskHandler.HandleAddTask))
	http.HandleFunc("/get", server.Auth(taskHandler.HandleGetTask))
	http.HandleFunc("/update", server.Auth(taskHandler.HandleUpdateTask))
	http.HandleFunc("/delete", server.Auth(taskHandler.HandleDeleteTask))

	go func() {
		server.StartServer(port)
	}()

	password := config.Config["TODO_PASSWORD"]
	token, err := getToken(password)
	if err != nil {
		log.Fatalf("Ошибка получения токена: %v", err)
	}
	log.Printf("Token: %s", token)

	select {}
}
