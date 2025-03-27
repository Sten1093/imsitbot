package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

// User - структура для хранения данных пользователя
type User struct {
	ID             int64  `json:"id"`
	State          string `json:"state"`
	Course         string `json:"course"`
	Group          string `json:"group"`
	Format         string `json:"format"`
	UserName       string `json:"user-name"`
	EducationLevel string `json:"education-level"`
}

// UserDAO - объект для работы с локальной БД
type UserDAO struct {
	db *sql.DB
}

// NewUserDAO - создаёт подключение к локальной БД
func NewUserDAO(dbPath string) *UserDAO {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	if db == nil {
		log.Fatal("Failed to open database connection")
	}

	// Создание таблицы, если она не существует
	createTableSQL := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		state TEXT,
		course TEXT,
		group_name TEXT,
		format TEXT,
		username TEXT,
		education_level TEXT
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatal(err)
	}

	return &UserDAO{db: db}
}

// SaveUser - сохраняет данные пользователя
func (dao *UserDAO) SaveUser(user *User) error {
	_, err := dao.db.Exec(`INSERT OR REPLACE INTO users (id, state, course, group_name, format, username, education_Level) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		user.ID, user.State, user.Course, user.Group, user.Format, user.UserName, user.EducationLevel)
	return err
}

// GetUser - загружает данные пользователя
func (dao *UserDAO) GetUser(userID int64) (*User, error) {
	var user User
	err := dao.db.QueryRow(`SELECT id, state, course, group_name, format,education_Level FROM users WHERE id = ?`, userID).Scan(&user.ID, &user.State, &user.Course, &user.Group, &user.Format, &user.EducationLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Пользователь не найден
		}
		return nil, err
	}
	return &user, nil
}

// DeleteUser - удаляет данные пользователя
func (dao *UserDAO) DeleteUser(userID int64) error {
	_, err := dao.db.Exec(`DELETE FROM users WHERE id = ?`, userID)
	return err
}

// GetAllUsers - получает список всех пользователей
func (dao *UserDAO) GetAllUsers() ([]*User, error) {
	rows, err := dao.db.Query("SELECT id, state, course, group_name, format, username, education_level FROM users")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var users []*User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.State, &user.Course, &user.Group, &user.Format, &user.UserName, &user.EducationLevel); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, nil
}

// Close - закрывает базу данных
func (dao *UserDAO) Close() {
	if err := dao.db.Close(); err != nil {
		log.Printf("Ошибка при закрытии базы данных: %v", err)
	} else {
		log.Println("База данных успешно закрыта")
	}
}
