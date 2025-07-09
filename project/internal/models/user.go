package models

type User struct {
	BaseModel
	Username     string   `json:"username"`
	PasswordHash string   `json:"-"`
	Email        string   `json:"email"`
	Phone        string   `json:"phone"`
	FullName     string   `json:"full_name"`
	UserType     UserType `json:"user_type"`
}

type UserType string

const (
	Resident UserType = "resident"
	Manager  UserType = "manager"
)
