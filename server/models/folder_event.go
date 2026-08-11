package models

import "gorm.io/gorm"

// FolderEvent represents the mapping between a folder and an event.
type FolderEvent struct {
	Id       ID      `json:"_id" bson:"_id,omitempty" gorm:"type:char(24)"`
	UserId   ID      `json:"userId" bson:"userId" gorm:"type:char(24);not null;uniqueIndex:idx_folder_event_user_event"`
	FolderId ID      `json:"folderId" bson:"folderId" gorm:"type:char(24);not null;index"`
	EventId  ID      `json:"eventId" bson:"eventId" gorm:"type:char(24);not null;uniqueIndex:idx_folder_event_user_event"`
	User     *User   `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Folder   *Folder `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Event    *Event  `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (f *FolderEvent) BeforeCreate(_ *gorm.DB) error {
	if f.Id.IsZero() {
		f.Id = NewID()
	}
	return nil
}
