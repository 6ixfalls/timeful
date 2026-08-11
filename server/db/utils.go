package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"schej.it/server/models"
	"schej.it/server/utils"
)

func GetFriendRequestById(requestID string) *models.FriendRequest {
	id, err := models.ParseID(requestID)
	if err != nil {
		return nil
	}
	var request models.FriendRequest
	if err := orm.First(&request, "id = ?", id).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			panic(err)
		}
		return nil
	}
	return &request
}

func DeleteFriendRequestById(requestID string) {
	id, err := models.ParseID(requestID)
	if err != nil {
		panic(err)
	}
	if err := orm.Delete(&models.FriendRequest{}, "id = ?", id).Error; err != nil {
		panic(err)
	}
}

func GetDailyUserLogByDate(date time.Time, timezoneOffset int) *models.DailyUserLog {
	offset, _ := time.ParseDuration(fmt.Sprintf("%dm", timezoneOffset))
	adjusted := date.Add(offset)
	start := utils.GetDateAtTime(adjusted, "00:00:00")
	end := utils.GetDateAtTime(adjusted, "23:59:59")
	var log models.DailyUserLog
	err := orm.Where("date >= ? AND date <= ?", start, end).First(&log).Error
	if err == gorm.ErrRecordNotFound {
		log.Date = models.NewDateTime(start)
		if err := orm.Create(&log).Error; err != nil {
			panic(err)
		}
	} else if err != nil {
		panic(err)
	}
	return &log
}

func UpdateDailyUserLog(user *models.User) {
	log := GetDailyUserLogByDate(time.Now(), user.TimezoneOffset)
	for _, id := range log.UserIds {
		if id == user.Id {
			return
		}
	}
	log.UserIds = append(log.UserIds, user.Id)
	if err := orm.Model(&models.DailyUserLog{}).Where("id = ?", log.Id).Update("user_ids", log.UserIds).Error; err != nil {
		panic(err)
	}
}
