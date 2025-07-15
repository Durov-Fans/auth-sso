package models

type User struct {
	ID             string `json:"id"`
	Hash           string `json:"hash"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Username       string `json:"username"`
	UserNameLocale string
	PhotoURL       string `json:"photo_url"`
	IsAdmin        bool   `json:"is_admin"`
}
