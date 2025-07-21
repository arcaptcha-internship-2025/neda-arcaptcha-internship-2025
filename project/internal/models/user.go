package models

type User struct {
	BaseModel
	Username string   `json:"username" db:"username"`
	Password string   `json:"password" db:"password"`
	Email    string   `json:"email" db:"email"`
	Phone    string   `json:"phone" db:"phone"`
	FullName string   `json:"full_name" db:"full_name"`
	UserType UserType `json:"user_type" db:"user_type"`
}

type UserType string

const (
	Resident UserType = "resident"
	Manager  UserType = "manager"
)
