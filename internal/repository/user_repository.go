package repository

import (
	"database/sql"
	"devsync/internal/model"
)

type UserRepository interface {
	GetByEmail(email string) (*model.User, error)
	GetByID(id uint) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	Create(user *model.User) error
	Search(query string) ([]model.User, error)
	UpdateProfile(user *model.User) error
	UpdateAvatar(userID uint, avatarPath string) error
}

type mysqlUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &mysqlUserRepository{db: db}
}

const userColumns = "id, name, username, email, password, avatar_path, created_at"

func scanUser(row interface {
	Scan(dest ...interface{}) error
}, u *model.User) error {
	return row.Scan(&u.ID, &u.Name, &u.Username, &u.Email, &u.Password, &u.AvatarPath, &u.CreatedAt)
}

func (r *mysqlUserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	query := "SELECT " + userColumns + " FROM users WHERE email = ?"
	if err := scanUser(r.db.QueryRow(query, email), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *mysqlUserRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	query := "SELECT " + userColumns + " FROM users WHERE id = ?"
	if err := scanUser(r.db.QueryRow(query, id), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *mysqlUserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	query := "SELECT " + userColumns + " FROM users WHERE username = ?"
	if err := scanUser(r.db.QueryRow(query, username), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *mysqlUserRepository) Create(user *model.User) error {
	query := "INSERT INTO users (name, email, password) VALUES (?, ?, ?)"
	res, err := r.db.Exec(query, user.Name, user.Email, user.Password)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	user.ID = uint(id)
	return nil
}

func (r *mysqlUserRepository) Search(query string) ([]model.User, error) {
	like := "%" + query + "%"
	sqlQuery := "SELECT " + userColumns + " FROM users WHERE name LIKE ? OR email LIKE ? ORDER BY name ASC LIMIT 20"
	rows, err := r.db.Query(sqlQuery, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []model.User{}
	for rows.Next() {
		var user model.User
		if err := scanUser(rows, &user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *mysqlUserRepository) UpdateProfile(user *model.User) error {
	_, err := r.db.Exec("UPDATE users SET name = ?, username = ? WHERE id = ?", user.Name, user.Username, user.ID)
	return err
}

func (r *mysqlUserRepository) UpdateAvatar(userID uint, avatarPath string) error {
	_, err := r.db.Exec("UPDATE users SET avatar_path = ? WHERE id = ?", avatarPath, userID)
	return err
}
