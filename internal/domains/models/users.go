package models

import "time"

type User struct {
	ID             string    `json:"id" sql:"id"`
	TgId           string    `json:"tgid" sql:"tgid"`
	FirstName      string    `json:"first_name" sql:"first_name"`
	LastName       string    `json:"last_name" sql:"last_name"`
	Username       string    `json:"user_name" sql:"user_name"`
	UserNameLocale string    `json:"user_name_locale" sql:"user_name_locale"`
	LastLogin      time.Time `json:"last_login" sql:"last_login"`
	PhotoURL       string    `json:"photo_url" sql:"photo_url"`
	IsAdmin        bool      `json:"is_admin" sql:"is_admin"`
	IsBanned       bool      `json:"is_banned" sql:"is_banned"`
}
type UserResponse struct {
	ID             string `json:"id" sql:"id"`
	TgId           string `json:"tgid" sql:"tgid"`
	FirstName      string `json:"first_name" sql:"first_name"`
	LastName       string `json:"last_name" sql:"last_name"`
	Username       string `json:"user_name" sql:"user_name"`
	UserNameLocale string `json:"user_name_locale" sql:"user_name_locale"`
	PhotoURL       string `json:"photo_url" sql:"photo_url"`
	IsBanned       bool   `json:"is_banned" sql:"is_banned"`
}
