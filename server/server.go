package server

import (
	"database/sql"
	"encoding/json"
	"final_sprint/api"
	"final_sprint/config"
	"final_sprint/utils"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handleAddTask(w, r)
	case http.MethodGet:
		api.HandleGetTask(w, r)
	case http.MethodPut:
		api.HandleUpdateTask(w, r)
	case http.MethodDelete:
		api.HandleDeleteTask(w, r)
	default:
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
	}
}

func handleAddTask(w http.ResponseWriter, r *http.Request) {
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
		task.Date = now.Format("20060102")
	}

	date, err := time.Parse("20060102", task.Date)
	if err != nil {
		http.Error(w, `{"error":"Дата представлена в неправильном формате"}`, http.StatusBadRequest)
		return
	}

	if task.Date < now.Format("20060102") && task.Repeat == "" {
		task.Date = now.Format("20060102")
	}

	if !date.After(now) {
		if task.Repeat == "" || date.Format("20060102") == now.Format("20060102") {
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

	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
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
			// Search is a text
			search = "%" + strings.ToLower(search) + "%"
			rows, err = db.Query(`
select t.id,
       t.date,
       upper(
               (WITH RECURSIVE under_name(test_text, char, level) as
                                   (select t.title, '', 0
                                    union
                                    select test_text, coalesce(lu.l, substr(test_text, level, 1)), under_name.level + 1
                                    from under_name
                                             left join (select 'А' as u, 'а' as l
                                                        union
                                                        select 'Б' as u, 'б' as l
                                                        union
                                                        select 'В' as u, 'в' as l
                                                        union
                                                        select 'Г' as u, 'г' as l
                                                        union
                                                        select 'Д' as u, 'д' as l
                                                        union
                                                        select 'Е' as u, 'е' as l
                                                        union
                                                        select 'Ё' as u, 'ё' as l
                                                        union
                                                        select 'Ж' as u, 'ж' as l
                                                        union
                                                        select 'З' as u, 'з' as l
                                                        union
                                                        select 'И' as u, 'и' as l
                                                        union
                                                        select 'Й' as u, 'й' as l
                                                        union
                                                        select 'К' as u, 'к' as l
                                                        union
                                                        select 'Л' as u, 'л' as l
                                                        union
                                                        select 'М' as u, 'м' as l
                                                        union
                                                        select 'Н' as u, 'н' as l
                                                        union
                                                        select 'О' as u, 'о' as l
                                                        union
                                                        select 'П' as u, 'п' as l
                                                        union
                                                        select 'Р' as u, 'р' as l
                                                        union
                                                        select 'С' as u, 'с' as l
                                                        union
                                                        select 'Т' as u, 'т' as l
                                                        union
                                                        select 'У' as u, 'у' as l
                                                        union
                                                        select 'Ф' as u, 'ф' as l
                                                        union
                                                        select 'Х' as u, 'х' as l
                                                        union
                                                        select 'Ц' as u, 'ц' as l
                                                        union
                                                        select 'Ч' as u, 'ч' as l
                                                        union
                                                        select 'Ш' as u, 'ш' as l
                                                        union
                                                        select 'Щ' as u, 'щ' as l
                                                        union
                                                        select 'Ь' as u, 'ь' as l
                                                        union
                                                        select 'Ы' as u, 'ы' as l
                                                        union
                                                        select 'Ъ' as u, 'ъ' as l
                                                        union
                                                        select 'Э' as u, 'э' as l
                                                        union
                                                        select 'Ю' as u, 'ю' as l
                                                        union
                                                        select 'Я' as u, 'я' as l) lu
                                                       on substr(test_text, level, 1) = lu.u
                                    where level <= length(test_text))
                select group_concat(char, '')
                from under_name)
       ) as lower_title,
       upper(
               (WITH RECURSIVE under_name(test_text, char, level) as
                                   (select t.comment, '', 0
                                    union
                                    select test_text, coalesce(lu.l, substr(test_text, level, 1)), under_name.level + 1
                                    from under_name
                                             left join (select 'А' as u, 'а' as l
                                                        union
                                                        select 'Б' as u, 'б' as l
                                                        union
                                                        select 'В' as u, 'в' as l
                                                        union
                                                        select 'Г' as u, 'г' as l
                                                        union
                                                        select 'Д' as u, 'д' as l
                                                        union
                                                        select 'Е' as u, 'е' as l
                                                        union
                                                        select 'Ё' as u, 'ё' as l
                                                        union
                                                        select 'Ж' as u, 'ж' as l
                                                        union
                                                        select 'З' as u, 'з' as l
                                                        union
                                                        select 'И' as u, 'и' as l
                                                        union
                                                        select 'Й' as u, 'й' as l
                                                        union
                                                        select 'К' as u, 'к' as l
                                                        union
                                                        select 'Л' as u, 'л' as l
                                                        union
                                                        select 'М' as u, 'м' as l
                                                        union
                                                        select 'Н' as u, 'н' as l
                                                        union
                                                        select 'О' as u, 'о' as l
                                                        union
                                                        select 'П' as u, 'п' as l
                                                        union
                                                        select 'Р' as u, 'р' as l
                                                        union
                                                        select 'С' as u, 'с' as l
                                                        union
                                                        select 'Т' as u, 'т' as l
                                                        union
                                                        select 'У' as u, 'у' as l
                                                        union
                                                        select 'Ф' as u, 'ф' as l
                                                        union
                                                        select 'Х' as u, 'х' as l
                                                        union
                                                        select 'Ц' as u, 'ц' as l
                                                        union
                                                        select 'Ч' as u, 'ч' as l
                                                        union
                                                        select 'Ш' as u, 'ш' as l
                                                        union
                                                        select 'Щ' as u, 'щ' as l
                                                        union
                                                        select 'Ь' as u, 'ь' as l
                                                        union
                                                        select 'Ы' as u, 'ы' as l
                                                        union
                                                        select 'Ъ' as u, 'ъ' as l
                                                        union
                                                        select 'Э' as u, 'э' as l
                                                        union
                                                        select 'Ю' as u, 'ю' as l
                                                        union
                                                        select 'Я' as u, 'я' as l) lu
                                                       on substr(test_text, level, 1) = lu.u
                                    where level <= length(test_text))
                select group_concat(char, '')
                from under_name)
       ) as lower_comment, 
    repeat
from scheduler t
where lower_title like ?
   or lower_comment like ?`, search, search)
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
	http.HandleFunc("/api/task", TaskHandler)
	http.HandleFunc("/api/tasks", TasksHandler)
	http.HandleFunc("/api/task/done", handleTaskDone)

	log.Printf("Сервер запущен на порту %d", port)
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
