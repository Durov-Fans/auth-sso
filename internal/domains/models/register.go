package models

type RegisterRequest struct {
	UserHash       string `json:"initData"`
	UserNameLocale string `json:"userNameLocale"`
	ServiceID      int64  `json:"serviceId"`
}
