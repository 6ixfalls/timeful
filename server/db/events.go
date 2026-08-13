package db

import (
	"math/rand"
	"time"

	"gorm.io/gorm"
	"schej.it/server/models"
)

func GetEventById(eventID string) *models.Event {
	id, err := models.ParseID(eventID)
	if err != nil {
		return nil
	}
	var event models.Event
	if err := orm.Where("id = ? AND COALESCE(is_deleted, false) = false", id).First(&event).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			panic(err)
		}
		return nil
	}
	return &event
}

func GetEventByShortId(shortID string) *models.Event {
	var event models.Event
	if err := orm.Where("short_id = ? AND COALESCE(is_deleted, false) = false", shortID).First(&event).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			panic(err)
		}
		return nil
	}
	return &event
}

func GetEventByEitherId(id string) *models.Event {
	if len(id) <= 10 {
		return GetEventByShortId(id)
	}
	return GetEventById(id)
}

func GetEventResponses(eventID string) []models.EventResponse {
	id, err := models.ParseID(eventID)
	if err != nil {
		return []models.EventResponse{}
	}
	var responses []models.EventResponse
	if err := orm.Where("event_id = ?", id).Find(&responses).Error; err != nil {
		panic(err)
	}
	return responses
}

func GetAttendees(eventID string) []models.Attendee {
	id, err := models.ParseID(eventID)
	if err != nil {
		return []models.Attendee{}
	}
	var attendees []models.Attendee
	if err := orm.Where("event_id = ?", id).Find(&attendees).Error; err != nil {
		panic(err)
	}
	return attendees
}

func GetEventsCreatedThisMonth(userID models.ID) int {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	var count int64
	if err := orm.Model(&models.Event{}).Where("owner_id = ? AND id >= ?", userID, models.NewIDFromTimestamp(start)).Count(&count).Error; err != nil {
		panic(err)
	}
	return int(count)
}

func GenerateShortEventId(eventID models.ID) string {
	random := rand.New(rand.NewSource(eventID.Timestamp().Unix()))
	letters := "23456789ABCDEFabcdef"
	id := ""
	for i := 0; i < 5; i++ {
		position := random.Intn(len(letters))
		id += letters[position : position+1]
	}
	for i := 0; GetEventByShortId(id) != nil && i < 5; i++ {
		position := random.Intn(len(letters))
		id += letters[position : position+1]
	}
	if GetEventByShortId(id) != nil {
		panic("couldn't generate unique event ID")
	}
	return id
}

func UpdateGuestResponseName(eventID, oldName, newName string) {
	id, err := models.ParseID(eventID)
	if err != nil {
		return
	}
	var response models.EventResponse
	if err := orm.Where("event_id = ? AND user_id = ?", id, oldName).First(&response).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return
		}
		panic(err)
	}
	response.UserId = newName
	if response.Response != nil {
		response.Response.Name = newName
	}
	if err := Updates(orm.Model(&models.EventResponse{}).Where("id = ?", response.Id), map[string]interface{}{
		"user_id":  response.UserId,
		"response": response.Response,
	}).Error; err != nil {
		panic(err)
	}
}

func GuestNameExists(eventID, guestName string) bool {
	event := GetEventByEitherId(eventID)
	if event == nil {
		return false
	}
	if id, err := models.ParseID(guestName); err == nil && GetUserById(id.Hex()) != nil {
		return true
	}
	var count int64
	if err := orm.Model(&models.EventResponse{}).Where("event_id = ? AND user_id = ?", event.Id, guestName).Count(&count).Error; err != nil {
		panic(err)
	}
	_, isID := models.ParseID(guestName)
	return count > 0 && isID != nil
}
