package models

import "gorm.io/gorm"

type Folder struct {
	Id     ID `json:"_id" bson:"_id,omitempty" gorm:"type:char(24)"`
	UserId ID `json:"userId" bson:"userId" gorm:"type:char(24);not null;index"`

	Name      string  `json:"name,omitempty" bson:"name,omitempty"`
	Color     *string `json:"color,omitempty" bson:"color,omitempty"`
	IsDeleted *bool   `json:"isDeleted,omitempty" bson:"isDeleted,omitempty" gorm:"index"`

	EventIds []ID  `json:"eventIds" bson:"-" gorm:"-"`
	User     *User `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (f *Folder) BeforeCreate(_ *gorm.DB) error {
	if f.Id.IsZero() {
		f.Id = NewID()
	}
	return nil
}
