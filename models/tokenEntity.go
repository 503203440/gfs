package models

type TokenEntity struct {
	ID          int64      `json:"id" gorm:"primaryKey;"`
	// 增加唯一索引与字段长度限制，避免高频 token 鉴权时的全表扫描
	Token       string     `json:"token" gorm:"uniqueIndex;size:64"`
	ClientID    int        `json:"clientId" gorm:"index"`
	Timestamp   string     `json:"timestamp"`
	ExpiresTime CustomTime `json:"expiresTime"`
	Used        bool       `json:"used" gorm:"index"`
}

func (TokenEntity) TableName() string {
	return "token_entity"
}
