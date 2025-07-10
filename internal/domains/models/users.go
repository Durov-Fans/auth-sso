package models

type User struct {
	ID             int64  `json:"id"`
	Hash           string `json:"hash"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Username       string `json:"username"`
	UserNameLocale string
	PhotoURL       string `json:"photo_url"`
}
