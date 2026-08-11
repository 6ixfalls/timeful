package db

import (
	"strings"

	"gorm.io/gorm"
	"schej.it/server/models"
)

func GetUserById(userID string) *models.User {
	id, err := models.ParseID(userID)
	if err != nil {
		return nil
	}
	var user models.User
	if err := orm.First(&user, "id = ?", id).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			panic(err)
		}
		return nil
	}
	return &user
}

func GetUserByStripeCustomerId(customerID string) *models.User {
	var user models.User
	if err := orm.First(&user, "stripe_customer_id = ?", customerID).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			panic(err)
		}
		return nil
	}
	return &user
}

func GetUserByEmail(email string) *models.User {
	email = strings.TrimSpace(email)
	if email == "" {
		return nil
	}
	var user models.User
	if err := orm.Where("LOWER(email) = LOWER(?)", email).First(&user).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			panic(err)
		}
		return nil
	}
	return &user
}
