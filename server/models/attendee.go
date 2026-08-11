package models

import "gorm.io/gorm"

type Attendee struct {
	Id      ID `json:"_id" bson:"_id,omitempty" gorm:"type:char(24)"`
	EventId ID `json:"eventId" bson:"eventId,omitempty" gorm:"type:char(24);not null;index"`

	Email    string `json:"email" bson:"email,omitempty" gorm:"not null"`
	Declined *bool  `json:"declined" bson:"declined,omitempty"`
	Event    *Event `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (a *Attendee) BeforeCreate(_ *gorm.DB) error {
	if a.Id.IsZero() {
		a.Id = NewID()
	}
	return nil
}
