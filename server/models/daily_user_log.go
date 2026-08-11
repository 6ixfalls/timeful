package models

import "gorm.io/gorm"

type DailyUserLog struct {
	Id      ID       `json:"_id" bson:"_id,omitempty" gorm:"type:char(24)"`
	Date    DateTime `json:"date" bson:"date,omitempty" gorm:"type:timestamptz;not null;uniqueIndex"`
	UserIds []ID     `json:"-" bson:"userIds" gorm:"serializer:json;type:jsonb"`
	Users   []User   `json:"users" bson:",omitempty" gorm:"-"`
}

func (l *DailyUserLog) BeforeCreate(_ *gorm.DB) error {
	if l.Id.IsZero() {
		l.Id = NewID()
	}
	return nil
}
