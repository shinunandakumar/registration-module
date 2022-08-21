package models

type User struct {
	Id       int    `json:"id,omitempty"`
	Username string `json:"user_name,omitempty" validate:"required"`
	Email    string `json:"email,omitempty" validate:"required"`
	Password string `json:"password,omitempty" validate:"required"`
}

type UserProfile struct {
	Id       int    `json:"id,omitempty"`
	Username string `json:"user_name,omitempty" `
	Email    string `json:"email,omitempty" `
	Name     string `json:"name,omitempty" `
	Phone    string `json:"telephone,omitempty" `
}
