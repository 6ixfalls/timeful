package models

import "gorm.io/gorm"

// User is persisted in PostgreSQL through GORM.
type User struct {
	TimezoneOffset int `json:"timezoneOffset" bson:"timezoneOffset"`

	// Profile info
	Id        ID     `json:"_id" bson:"_id,omitempty" gorm:"type:char(24)"`
	Email     string `json:"email" bson:"email,omitempty" gorm:"uniqueIndex;not null"`
	FirstName string             `json:"firstName" bson:"firstName,omitempty"`
	LastName  string             `json:"lastName" bson:"lastName,omitempty"`
	Picture   string             `json:"picture" bson:"picture,omitempty"`

	// Whether the user has set a custom name for themselves, i.e. don't change their name when they sign in
	HasCustomName *bool `json:"hasCustomName" bson:"hasCustomName,omitempty"`

	// CalendarAccounts is a mapping from {`email_CALENDARTYPE` => CalendarAccount} that contains all the
	// additional accounts the user wants to see google calendar events for
	CalendarAccounts map[string]CalendarAccount `json:"calendarAccounts" bson:"calendarAccounts,omitempty" gorm:"serializer:storage_json;type:jsonb"`

	// The calendarAccountKey of the account the user first signed in with
	PrimaryAccountKey *string `json:"primaryAccountKey" bson:"primaryAccountKey,omitempty"`

	// Google OAuth stuff
	TokenOrigin TokenOriginType `json:"-" bson:"tokenOrigin,omitempty"`

	// Calendar options
	CalendarOptions *CalendarOptions `json:"calendarOptions" bson:"calendarOptions,omitempty" gorm:"serializer:json;type:jsonb"`

	// Stripe customer ID
	StripeCustomerId *string `json:"stripeCustomerId" bson:"stripeCustomerId,omitempty" gorm:"index"`
	IsPremium        *bool   `json:"isPremium" bson:"isPremium,omitempty"`
	NumEventsCreated int     `json:"numEventsCreated" bson:"numEventsCreated,omitempty"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.Id.IsZero() {
		u.Id = NewID()
	}
	return nil
}

// Declare the possible types of TokenOrigin
type TokenOriginType string

const (
	Undefined TokenOriginType = ""
	IOS       TokenOriginType = "ios"
	ANDROID   TokenOriginType = "android"
	WEB       TokenOriginType = "web"
)

type UserStatus string

const (
	FREE UserStatus = "free"
	BUSY UserStatus = "busy"
)
