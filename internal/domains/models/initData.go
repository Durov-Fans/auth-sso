package models

type InitDataRequest struct {
	InitData  string `json:"initData"`
	ServiceId int64  `json:"serviceId"`
}
type InitDataUnsafe struct {
	User         TelegramUser `json:"user"`
	ChatInstance string       `json:"chat_instance"`
	ChatType     string       `json:"chat_type"`
	AuthDate     string       `json:"auth_date"`
	Signature    string       `json:"signature"`
	Hash         string       `json:"hash"`
}
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
}
