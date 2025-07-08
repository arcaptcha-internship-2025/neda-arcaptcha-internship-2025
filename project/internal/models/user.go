package models

type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	FullName     string `json:"full_name"`
	UserType     string `json:"user_type"` // e.g., 'resident', 'manager'
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
