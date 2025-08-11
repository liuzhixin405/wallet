package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint64         `json:"id" gorm:"primaryKey;autoIncrement"`
	Username    string         `json:"username" gorm:"type:varchar(50);not null;uniqueIndex"`
	Password    string         `json:"-" gorm:"type:varchar(255);not null"`
	Email       string         `json:"email" gorm:"type:varchar(100);not null;uniqueIndex"`
	Status      bool           `json:"status" gorm:"not null;default:true"`
	CreatedTime time.Time      `json:"created_time" gorm:"not null;autoCreateTime"`
	UpdatedTime time.Time      `json:"updated_time" gorm:"not null;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (User) TableName() string {
	return "users"
}
