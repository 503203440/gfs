package models

type ClientInfoEntity struct {
	Id        *int `gorm:"primarykey"`
	SecertKey string
}

func (ClientInfoEntity) TableName() string {
	return "client_info_entity"
}
