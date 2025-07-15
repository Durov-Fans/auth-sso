package models

type RegisterRequest struct {
	UserHash       string `json:"userHash"`
	UserNameLocale string `json:"userNameLocale"`
	ServiceID      int64  `json:"serviceId"`
}
