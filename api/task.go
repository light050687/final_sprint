package api

import (
	"database/sql"
	"encoding/json"
	"final_sprint/config"
	"final_sprint/utils"
	"net/http"
	"time"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func HandleGetTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	dbFile := config.GetDBFile()
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		http.Error(w, `{"error":"Ошибка подключения соединения"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var task Task
	err = db.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"Ошибка запроса к базе данных"}`, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(task)
}

func HandleUpdateTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Ошибка декодирования JSON", http.StatusBadRequest)
		return
	}

	if task.ID == "" {
		http.Error(w, `{"error":"Не указан идентификатор задачи"}`, http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		http.Error(w, `{"error":"Не указан заголовок задачи"}`, http.StatusBadRequest)
		return
	}

	now := time.Now()
	if task.Date == "" {
		task.Date = now.Format("20060102")
	}

	date, err := time.Parse("20060102", task.Date)
	if err != nil {
		http.Error(w, `{"error":"Дата представлена в неправильном формате"}`, http.StatusBadRequest)
		return
	}

	if !date.After(now) {
		if task.Repeat == "" {
			task.Date = now.Format("20060102")
		} else {
			nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				http.Error(w, `{"error":"Неправильный формат правила повторения"}`, http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		}
	} else if task.Repeat != "" && date.Format("20060102") == now.Format("20060102") {
		nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			http.Error(w, `{"error":"Неправильный формат правила повторения"}`, http.StatusBadRequest)
			return
		}
		task.Date = nextDate
	}

	dbFile := config.GetDBFile()
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		http.Error(w, "Ошибка подключения к базе данных", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		http.Error(w, "Ошибка обновления базы данных", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, "Ошибка получения количества затронутых строк", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]interface{}{})
}

func HandleDeleteTask(w http.ResponseWriter, r *http.Request) {
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

	var task Task
	err = db.QueryRow(`SELECT id FROM scheduler WHERE id = ?`, id).Scan(&task.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"Ошибка запроса к базе данных"}`, http.StatusInternalServerError)
		}
		return
	}

	_, err = db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		http.Error(w, `{"error":"Ошибка удаления из базы данных"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]interface{}{})
}
