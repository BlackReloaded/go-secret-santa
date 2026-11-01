package secretsanta

import (
	"database/sql"
	"fmt"
)

// User is a represatation of participant
type User struct {
	ID        uint32
	Firstname string
	Lastname  string
	Enabled   bool
	Email     string
}

// UserSlice is a type to sort a year slices
type UserSlice struct {
	Users []*User
}

func (us UserSlice) Len() int {
	return len(us.Users)
}

func (us UserSlice) Less(i, j int) bool {
	return us.Users[i].Lastname < us.Users[j].Lastname
}

func (us UserSlice) Swap(i, j int) {
	us.Users[i], us.Users[j] = us.Users[j], us.Users[i]
}

func (ss *SecretSanta) initUserTable() error {
	_, err := ss.db.Exec(`CREATE TABLE IF NOT EXISTS User (
		    id INTEGER PRIMARY KEY,
		    firstname VARCHAR(255),
			lastname VARCHAR(255) NOT NULL,
			enabled BOOLEAN DEFAULT true,
			email VARCHAR(255) unique
		);`)
	return err
}

// AddUser add a user to the database
func (ss *SecretSanta) AddUser(user *User) (uint32, error) {
	if user == nil {
		return 0, fmt.Errorf("user is nil")
	}
	result, err := ss.db.Exec(`INSERT INTO User(firstname,lastname,enabled,email) 
		VALUES(?,?,?,?)`,
		user.Firstname, user.Lastname, user.Enabled, user.Email)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}
	user.ID = uint32(id)
	return user.ID, nil
}

// UpdateUser updates a user in the database
func (ss *SecretSanta) UpdateUser(user *User) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}
	if user.ID == 0 {
		return fmt.Errorf("user id is 0")
	}
	result, err := ss.db.Exec(`UPDATE User
		SET firstname = ?,lastname = ?,enabled = ?,email = ?
		WHERE ID = ?`,
		user.Firstname, user.Lastname, user.Enabled, user.Email, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("failed to update user")
	}
	return nil
}

// ListUsers lists users in the database. With disabled=false only enabled
// users are returned, with disabled=true only disabled users.
func (ss *SecretSanta) ListUsers(disabled bool) ([]*User, error) {
	rows, err := ss.db.Query("SELECT * FROM User WHERE enabled = ?;", !disabled)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()
	return scanUsers(rows)
}

// ListAllUsers lists all users in the database, enabled and disabled
func (ss *SecretSanta) ListAllUsers() ([]*User, error) {
	rows, err := ss.db.Query("SELECT * FROM User;")
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()
	return scanUsers(rows)
}

func scanUsers(rows *sql.Rows) ([]*User, error) {
	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(&user.ID, &user.Firstname, &user.Lastname, &user.Enabled, &user.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUser gets a user by id
func (ss *SecretSanta) GetUser(id uint32) (*User, error) {
	row := ss.db.QueryRow("SELECT * FROM User WHERE ID = ?;", id)
	user := &User{}
	err := row.Scan(&user.ID, &user.Firstname, &user.Lastname, &user.Enabled, &user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to scan user row: %w", err)
	}
	return user, nil
}

// RemoveUser remove a user by id
func (ss *SecretSanta) RemoveUser(id uint32) error {
	result, err := ss.db.Exec("DELETE FROM User WHERE ID = ?;", id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read affected rows: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("failed to delete user")
	}
	return nil
}
