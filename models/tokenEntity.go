package models

type TokenEntity struct {
	ID          int64      `json:"id" gorm:"primaryKey;"`
	Token       string     `json:"token"`
	ClientID    int        `json:"clientId"`
	Timestamp   string     `json:"timestamp"`
	ExpiresTime CustomTime `json:"expiresTime"`
	Used        bool       `json:"used"`
}

func (TokenEntity) TableName() string {
	return "token_entity"
}
