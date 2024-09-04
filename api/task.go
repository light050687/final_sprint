package api

import (
	"database/sql"
	"encoding/json"
	"final_sprint/utils"
	"net/http"
	"time"
)

const DateFormat = "20060102"

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TaskHandler struct {
	db *sql.DB
}

func NewTaskHandler(db *sql.DB) *TaskHandler {
	return &TaskHandler{db: db}
}

func (h *TaskHandler) HandleAddTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, "Ошибка декодирования JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		http.Error(w, `{"error":"Не указан заголовок задачи"}`, http.StatusBadRequest)
		return
	}

	now := time.Now()
	if task.Date == "" {
		task.Date = now.Format(DateFormat)
	}

	date, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		http.Error(w, `{"error":"Дата представлена в неправильном формате"}`, http.StatusBadRequest)
		return
	}

	if task.Date < now.Format(DateFormat) && task.Repeat == "" {
		task.Date = now.Format(DateFormat)
	}

	if !date.After(now) {
		if task.Repeat == "" || date.Format(DateFormat) == now.Format(DateFormat) {
			task.Date = now.Format(DateFormat)
		} else {
			nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				http.Error(w, `{"error":"Неправильный формат правила повторения"}`, http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		}
	} else if task.Repeat != "" && date.Format(DateFormat) == now.Format(DateFormat) {
		nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			http.Error(w, `{"error":"Неправильный формат правила повторения"}`, http.StatusBadRequest)
			return
		}
		task.Date = nextDate
	}

	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := h.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, "Ошибка вставки в базу данных", http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Ошибка получения последнего вставленного идентификатора", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
}

func (h *TaskHandler) HandleGetTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	var task Task
	err := h.db.QueryRow(`SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`, id).Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
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

func (h *TaskHandler) HandleUpdateTask(w http.ResponseWriter, r *http.Request) {
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
		task.Date = now.Format(DateFormat)
	}

	date, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		http.Error(w, `{"error":"Дата представлена в неправильном формате"}`, http.StatusBadRequest)
		return
	}

	if !date.After(now) {
		if task.Repeat == "" {
			task.Date = now.Format(DateFormat)
		} else {
			nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				http.Error(w, `{"error":"Неправильный формат правила повторения"}`, http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		}
	} else if task.Repeat != "" && date.Format(DateFormat) == now.Format(DateFormat) {
		nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			http.Error(w, `{"error":"Неправильный формат правила повторения"}`, http.StatusBadRequest)
			return
		}
		task.Date = nextDate
	}

	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := h.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
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

func (h *TaskHandler) HandleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
		return
	}

	res, err := h.db.Exec(`DELETE FROM scheduler WHERE id = ?`, id)
	if err != nil {
		http.Error(w, `{"error":"Ошибка удаления из базы данных"}`, http.StatusInternalServerError)
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

func (h *TaskHandler) TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.HandleAddTask(w, r)
	case http.MethodGet:
		h.HandleGetTask(w, r)
	case http.MethodPut:
		h.HandleUpdateTask(w, r)
	case http.MethodDelete:
		h.HandleDeleteTask(w, r)
	default:
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
	}
}
