package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const Schema = `CREATE TABLE IF NOT EXISTS scheduler (
   				id INTEGER PRIMARY KEY AUTOINCREMENT,
    			date CHAR(8) NOT NULL DEFAULT "",
    			title VARCHAR(100) NOT NULL DEFAULT "",
    			comment TEXT,
    			repeat VARCHAR(128) NOT NULL DEFAULT ""
  				);
  				CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler (date);
				`

type Base struct {
	Conn *sql.DB
}

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func Init(dbFile string) (*Base, error) {
	envFile := os.Getenv("TODO_DBFILE")
	if envFile != "" {
		dbFile = envFile
	}

	path, err := dbPath(dbFile)
	if err != nil {
		return nil, fmt.Errorf(" Ошибка строения путя к Бд %v", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("Ошибка открытия бд %v", err)
	}

	_, err = db.Exec(Schema)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("Ошибка заполнения бд %v", err)
	}

	return &Base{
			Conn: db,
		},
		nil
}

func dbPath(dbFile string) (string, error) {
	path, err := filepath.Abs(dbFile)
	if err != nil {
		return "", err
	}

	return path, err
}

func (s *Base) AddTask(task *Task) (int64, error) {
	var id int64
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := s.Conn.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}

	id, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Base) Tasks(limit int) ([]*Task, error) {
	tasks := make([]*Task, 0, 50)
	query := `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date LIMIT ?`
	rows, err := s.Conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		err = rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *Base) SearchTasksParam(param string, limit int) ([]*Task, error) {
	tasks := make([]*Task, 0, 50)
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?`

	rows, err := s.Conn.Query(query, "%"+param+"%", "%"+param+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *Base) SearchTasksDate(date string, limit int) ([]*Task, error) {
	tasks := make([]*Task, 0, 50)
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? LIMIT ?`
	rows, err := s.Conn.Query(query, date, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}
