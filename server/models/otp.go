package models

import (
	"gorm.io/gorm"
	"time"
)

type OtpCode struct {
	Id        ID        `json:"_id" bson:"_id,omitempty" gorm:"type:char(24)"`
	Email     string    `json:"email" bson:"email" gorm:"not null;index"`
	Code      string    `json:"-" bson:"code" gorm:"not null"`
	ExpiresAt time.Time `json:"-" bson:"expiresAt" gorm:"not null;index"`
	Attempts  int                `json:"-" bson:"attempts"`
}

func (o *OtpCode) BeforeCreate(_ *gorm.DB) error {
	if o.Id.IsZero() {
		o.Id = NewID()
	}
	return nil
}
