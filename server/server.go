package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"final_sprint/api"
	"final_sprint/config"
	"final_sprint/utils"

	_ "modernc.org/sqlite"
)

const limit = 50

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeat := r.URL.Query().Get("repeat")

	if nowStr == "" || dateStr == "" || repeat == "" {
		http.Error(w, `{"error":"Отсутствуют обязательные параметры"}`, http.StatusBadRequest)
		return
	}

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, `{"error":"Неправильный формат текущей даты"}`, http.StatusBadRequest)
		return
	}

	nextDate, err := utils.NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.Write([]byte(nextDate))
}

func TasksHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	dbFile := config.GetDBFile()
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		http.Error(w, `{"error":"Ошибка подключения к базе данных"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var rows *sql.Rows
	if search == "" {
		rows, err = db.Query(`SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT 50`)
	} else {
		search = strings.TrimSpace(search)
		if date, err := time.Parse("02.01.2006", search); err == nil {
			// Search is a date
			rows, err = db.Query(`SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date LIMIT 50`, date.Format("20060102"))
		} else {
			search = "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
			rows, err = db.Query(`SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ? COLLATE NOCASE`, search, search, limit)
		}
	}
	if err != nil {
		http.Error(w, `{"error":"Ошибка запроса к базе данных"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []map[string]string
	for rows.Next() {
		var id, date, title, comment, repeat string
		if err := rows.Scan(&id, &date, &title, &comment, &repeat); err != nil {
			http.Error(w, `{"error":"Ошибка сканирования результатов базы данных"}`, http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, map[string]string{
			"id":      id,
			"date":    date,
			"title":   title,
			"comment": comment,
			"repeat":  repeat,
		})
	}

	if tasks == nil {
		tasks = []map[string]string{}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
}

func handleTaskDone(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	dbFile := config.GetDBFile()
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		http.Error(w, `{"error":"Ошибка подключения к базе данных"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var task api.Task
	err = db.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"Ошибка запроса к базе данных"}`, http.StatusInternalServerError)
		}
		return
	}

	if task.Repeat == "" {
		_, err = db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
		if err != nil {
			http.Error(w, `{"error":"Ошибка удаления из базы данных"}`, http.StatusInternalServerError)
			return
		}
	} else {
		now := time.Now()
		nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		_, err = db.Exec(`UPDATE scheduler SET date = ? WHERE id = ?`, nextDate, id)
		if err != nil {
			http.Error(w, `{"error":"Ошибка обновления базы данных"}`, http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

func StartServer(port int) {
	webDir := "./web"
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/signin", SignInHandler)
	http.HandleFunc("/api/nextdate", NextDateHandler)

	// Создание экземпляра TaskHandler
	dbFile := config.GetDBFile()
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	taskHandler := api.NewTaskHandler(db)
	http.HandleFunc("/api/task", taskHandler.TaskHandler)
	http.HandleFunc("/api/tasks", TasksHandler)
	http.HandleFunc("/api/task/done", handleTaskDone)

	log.Printf("Сервер запущен на порту %d", port)
	err = http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
