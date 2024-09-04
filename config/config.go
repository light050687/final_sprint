package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
)

var Config map[string]string

func InitConfig() {
	Config = make(map[string]string)
	Config["TODO_PASSWORD"] = os.Getenv("TODO_PASSWORD")
}

func GetPort() int {
	port := 7540
	envPort := os.Getenv("TODO_PORT")
	if len(envPort) > 0 {
		if eport, err := strconv.Atoi(envPort); err == nil {
			port = eport
		}
	}
	return port
}

func GetDBFile() string {
	dbFile := "scheduler.db"
	envDBFile := os.Getenv("TODO_DBFILE")
	if len(envDBFile) > 0 {
		dbFile = envDBFile
	} else {
		projectRoot, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		dbFile = filepath.Join(projectRoot, dbFile)
	}
	return dbFile
}
