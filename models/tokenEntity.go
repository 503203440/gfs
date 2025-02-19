package models

type TokenEntity struct {
	ID          int64      `json:"id" gorm:"primaryKey;"`
	Token       string     `json:"token"`
	ClientID    int        `json:"client_id"`
	Timestamp   string     `json:"timestamp"`
	ExpiresTime CustomTime `json:"expires_time"`
	Used        bool       `json:"used"`
}

func (TokenEntity) TableName() string {
	return "token_entity"
}
