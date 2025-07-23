package models

type User struct {
	BaseModel
	Username       string   `json:"username" db:"username"`
	Password       string   `json:"password" db:"password"`
	Email          string   `json:"email" db:"email"`
	Phone          string   `json:"phone" db:"phone"`
	FullName       string   `json:"full_name" db:"full_name"`
	UserType       UserType `json:"user_type" db:"user_type"`
	TelegramUser   string   `json:"telegram_user" db:"telegram_user"`       // telegram username without @
	TelegramChatID int64    `json:"telegram_chat_id" db:"telegram_chat_id"` // will be set after user starts the bot
}

type UserType string

const (
	Resident UserType = "resident"
	Manager  UserType = "manager"
)
