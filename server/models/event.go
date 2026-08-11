package models

import "gorm.io/gorm"

type EventType string

const (
	SPECIFIC_DATES EventType = "specific_dates"
	DOW            EventType = "dow"
	GROUP          EventType = "group"
)

// Object containing information associated with the remindee
type Remindee struct {
	Email     string   `json:"email" bson:"email,omitempty"`
	TaskIds   []string `json:"-" bson:"taskIds,omitempty"` // Task IDs of the scheduled emails
	Responded *bool    `json:"responded" bson:"responded,omitempty"`
}

type SignUpBlock struct {
	Id        ID        `json:"_id" bson:"_id,omitempty"`
	Name      string              `json:"name" bson:"name,omitempty"`
	Capacity  *int                `json:"capacity" bson:"capacity,omitempty"`
	StartDate *DateTime `json:"startDate" bson:"startDate,omitempty"`
	EndDate   *DateTime `json:"endDate" bson:"endDate,omitempty"`
}

type SignUpResponse struct {
	// The IDs of the sign up blocks that the user has signed up for
	SignUpBlockIds []ID `json:"signUpBlockIds" bson:"signUpBlockIds,omitempty"`

	// Guest information
	Name  string `json:"name" bson:"name,omitempty"`
	Email string `json:"email" bson:"email,omitempty"`

	// User information
	UserId ID    `json:"userId" bson:"userId,omitempty"`
	User   *User              `json:"user" bson:",omitempty"`
}

type Event struct {
	Id          ID      `json:"_id" bson:"_id,omitempty" gorm:"type:char(24)"`
	ShortId     *string `json:"shortId" bson:"shortId,omitempty" gorm:"uniqueIndex"`
	OwnerId     ID      `json:"ownerId" bson:"ownerId,omitempty" gorm:"type:char(24);index"`
	Name        string             `json:"name" bson:"name,omitempty"`
	Description *string            `json:"description" bson:"description,omitempty"`
	IsArchived  *bool   `json:"isArchived" bson:"isArchived,omitempty" gorm:"index"`
	IsDeleted   *bool   `json:"isDeleted" bson:"isDeleted,omitempty" gorm:"index"`

	Duration                 *float32             `json:"duration" bson:"duration,omitempty"`
	Dates                    []DateTime `json:"dates" bson:"dates,omitempty" gorm:"serializer:json;type:jsonb"`
	NotificationsEnabled     *bool                `json:"notificationsEnabled" bson:"notificationsEnabled,omitempty"`
	SendEmailAfterXResponses *int                 `json:"sendEmailAfterXResponses" bson:"sendEmailAfterXResponses,omitempty"`
	When2meetHref            *string              `json:"when2meetHref" bson:"when2meetHref,omitempty"`
	CollectEmails            *bool                `json:"collectEmails" bson:"collectEmails,omitempty"`
	TimeIncrement            *int                 `json:"timeIncrement" bson:"timeIncrement,omitempty"`

	// Used for specific times for specific dates feature
	HasSpecificTimes *bool                `json:"hasSpecificTimes" bson:"hasSpecificTimes,omitempty"`
	Times            []DateTime `json:"times" bson:"times,omitempty" gorm:"serializer:json;type:jsonb"`

	Type EventType `json:"type" bson:"type,omitempty"`

	// PostHog ID for the event creator
	CreatorPosthogId *string `json:"creatorPosthogId" bson:"creatorPosthogId,omitempty"`

	// Sign up form details
	IsSignUpForm    *bool                      `json:"isSignUpForm" bson:"isSignUpForm,omitempty"`
	SignUpBlocks    *[]SignUpBlock             `json:"signUpBlocks" bson:"signUpBlocks,omitempty" gorm:"serializer:json;type:jsonb"`
	SignUpResponses map[string]*SignUpResponse `json:"signUpResponses" bson:"signUpResponses" gorm:"serializer:json;type:jsonb"`

	// Whether to start the event on Monday (as opposed to Sunday, used for DOW events)
	StartOnMonday *bool `json:"startOnMonday" bson:"startOnMonday,omitempty"`

	// Whether to enable blind availability
	BlindAvailabilityEnabled *bool `json:"blindAvailabilityEnabled" bson:"blindAvailabilityEnabled,omitempty"`

	// Whether to only poll for days, not times
	DaysOnly *bool `json:"daysOnly" bson:"daysOnly,omitempty"`

	// Availability responses - old format for backward compatibility (fetched from eventResponses collection)
	ResponsesMap map[string]*Response `json:"responses" bson:"-" gorm:"-"`

	// Used to store the number of responses for the event
	NumResponses *int `json:"numResponses" bson:"numResponses,omitempty"`

	// Scheduled event
	ScheduledEvent  *CalendarEvent `json:"scheduledEvent" bson:"scheduledEvent,omitempty" gorm:"serializer:json;type:jsonb"`
	CalendarEventId string         `json:"calendarEventId" bson:"calendarEventId,omitempty"`

	// Remindees
	Remindees *[]Remindee `json:"remindees" bson:"remindees,omitempty" gorm:"serializer:storage_json;type:jsonb"`

	// Attendees for an availability group (fetched from Attendees collection)
	Attendees *[]Attendee `json:"attendees" bson:"-" gorm:"-"`

	// Whether the user has responded to the availability group (fetched based on whether user is in Attendees)
	HasResponded *bool `json:"hasResponded" bson:"-"`
}

func (e *Event) BeforeCreate(_ *gorm.DB) error {
	if e.Id.IsZero() {
		e.Id = NewID()
	}
	return nil
}

func (e *Event) GetId() string {
	if e.ShortId != nil {
		return *e.ShortId
	}

	return e.Id.Hex()
}
