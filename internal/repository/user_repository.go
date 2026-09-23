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
	SearchVisibleTo(userID uint, query string) ([]model.User, error)
	ListForAdmin(query string) ([]model.AdminUser, error)
	UpdateProfile(user *model.User) error
	UpdateAvatar(userID uint, avatarPath string) error
}

type mysqlUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &mysqlUserRepository{db: db}
}

const userColumns = "id, name, username, email, password, avatar_path, role, auth_provider, created_at"

func scanUser(row interface {
	Scan(dest ...interface{}) error
}, u *model.User) error {
	return row.Scan(&u.ID, &u.Name, &u.Username, &u.Email, &u.Password, &u.AvatarPath, &u.Role, &u.AuthProvider, &u.CreatedAt)
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
	if user.Role == "" {
		user.Role = "user"
	}
	if user.AuthProvider == "" {
		user.AuthProvider = "password"
	}
	query := "INSERT INTO users (name, email, password, role, auth_provider) VALUES (?, ?, ?, ?, ?)"
	res, err := r.db.Exec(query, user.Name, user.Email, user.Password, user.Role, user.AuthProvider)
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

// SearchVisibleTo is Search for non-admins: it only returns people who already
// share a workspace with userID, plus an exact email match — so typing a
// colleague's full address still finds them, but nobody can enumerate every
// registered email by searching for "@" or single letters.
func (r *mysqlUserRepository) SearchVisibleTo(userID uint, query string) ([]model.User, error) {
	like := "%" + query + "%"
	sqlQuery := "SELECT " + userColumns + ` FROM users u
		WHERE (u.name LIKE ? OR u.email LIKE ?)
		  AND (u.email = ? OR EXISTS (
			SELECT 1 FROM workspace_members mine
			JOIN workspace_members theirs ON theirs.workspace_id = mine.workspace_id
			WHERE mine.user_id = ? AND theirs.user_id = u.id))
		ORDER BY u.name ASC LIMIT 20`
	rows, err := r.db.Query(sqlQuery, like, like, query, userID)
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

// ListForAdmin returns every registered user (newest first), optionally
// filtered by name/email, each with the workspaces they belong to.
func (r *mysqlUserRepository) ListForAdmin(query string) ([]model.AdminUser, error) {
	like := "%" + query + "%"
	sqlQuery := "SELECT " + userColumns + " FROM users WHERE ? = '' OR name LIKE ? OR email LIKE ? ORDER BY created_at DESC, id DESC LIMIT 500"
	rows, err := r.db.Query(sqlQuery, query, like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []model.AdminUser{}
	index := map[uint]int{}
	for rows.Next() {
		var u model.AdminUser
		if err := scanUser(rows, &u.User); err != nil {
			return nil, err
		}
		u.Workspaces = []model.AdminUserWorkspace{}
		index[u.ID] = len(users)
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return users, nil
	}

	wsRows, err := r.db.Query(`SELECT wm.user_id, w.id, w.name, wm.role
		FROM workspace_members wm
		JOIN workspaces w ON w.id = wm.workspace_id
		ORDER BY w.name ASC`)
	if err != nil {
		return nil, err
	}
	defer wsRows.Close()
	for wsRows.Next() {
		var userID uint
		var ws model.AdminUserWorkspace
		if err := wsRows.Scan(&userID, &ws.ID, &ws.Name, &ws.Role); err != nil {
			return nil, err
		}
		if i, ok := index[userID]; ok {
			users[i].Workspaces = append(users[i].Workspaces, ws)
		}
	}
	return users, wsRows.Err()
}

func (r *mysqlUserRepository) UpdateProfile(user *model.User) error {
	_, err := r.db.Exec("UPDATE users SET name = ?, username = ? WHERE id = ?", user.Name, user.Username, user.ID)
	return err
}

func (r *mysqlUserRepository) UpdateAvatar(userID uint, avatarPath string) error {
	_, err := r.db.Exec("UPDATE users SET avatar_path = ? WHERE id = ?", avatarPath, userID)
	return err
}
