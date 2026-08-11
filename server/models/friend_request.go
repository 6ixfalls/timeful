package models

import "gorm.io/gorm"

type FriendRequest struct {
	Id        ID       `json:"_id" bson:"_id,omitempty" gorm:"type:char(24)"`
	From      ID       `json:"from" bson:"from,omitempty" gorm:"column:from_id;type:char(24);not null;index"`
	FromUser  *User    `json:"fromUser" bson:",omitempty" gorm:"foreignKey:From;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	To        ID       `json:"to" bson:"to,omitempty" gorm:"column:to_id;type:char(24);not null;index"`
	ToUser    *User    `json:"toUser" bson:",omitempty" gorm:"foreignKey:To;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedAt DateTime `json:"createdAt" bson:"createdAt,omitempty" gorm:"type:timestamptz;not null"`
}

func (f *FriendRequest) BeforeCreate(_ *gorm.DB) error {
	if f.Id.IsZero() {
		f.Id = NewID()
	}
	if f.CreatedAt == 0 {
		f.CreatedAt = NewDateTime(f.Id.Timestamp())
	}
	return nil
}
