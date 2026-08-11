package models

import "gorm.io/gorm"

type EventResponse struct {
	Id      ID `json:"id" bson:"_id,omitempty" gorm:"type:char(24)"`
	EventId ID `json:"eventId" bson:"eventId" gorm:"type:char(24);not null;index"`

	UserId   string    `json:"userId" bson:"userId" gorm:"not null"`
	Response *Response `json:"response" bson:"response" gorm:"serializer:json;type:jsonb;not null"`
	Event    *Event    `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (r *EventResponse) BeforeCreate(_ *gorm.DB) error {
	if r.Id.IsZero() {
		r.Id = NewID()
	}
	return nil
}

// A response object containing an array of times that the given user is available
type Response struct {
	// Guest information
	Name  string `json:"name" bson:"name,omitempty"`
	Email string `json:"email" bson:"email,omitempty"`

	// User information
	UserId ID    `json:"userId" bson:"userId,omitempty"`
	User   *User              `json:"user" bson:",omitempty"`

	// Availability
	Availability []DateTime `json:"availability" bson:"availability"`
	IfNeeded     []DateTime `json:"ifNeeded" bson:"ifNeeded"`

	// Mapping from the start date of a day to the available times for that day
	ManualAvailability *map[DateTime][]DateTime `json:"manualAvailability" bson:"manualAvailability,omitempty"`

	// Calendar availability variables for Availability Groups feature
	UseCalendarAvailability *bool                `json:"useCalendarAvailability" bson:"useCalendarAvailability,omitempty"`
	EnabledCalendars        *map[string][]string `json:"enabledCalendars" bson:"enabledCalendars,omitempty"` // Maps email to an array of sub calendar ids
	CalendarOptions         *CalendarOptions     `json:"calendarOptions" bson:"calendarOptions,omitempty"`
}
