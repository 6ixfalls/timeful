package db

import (
	"gorm.io/gorm"
	"schej.it/server/models"
)

func CreateFolder(folder *models.Folder) (models.ID, error) {
	err := orm.Create(folder).Error
	return folder.Id, err
}

func GetFolderById(folderID, userID models.ID) (*models.Folder, error) {
	var folder models.Folder
	err := orm.Where("id = ? AND user_id = ? AND COALESCE(is_deleted, false) = false", folderID, userID).First(&folder).Error
	return &folder, err
}

func GetAllFolders(userID models.ID) ([]models.Folder, error) {
	var folders []models.Folder
	if err := orm.Where("user_id = ? AND COALESCE(is_deleted, false) = false", userID).Find(&folders).Error; err != nil {
		return nil, err
	}
	for i := range folders {
		events, err := GetEventsInFolder(folders[i].Id, userID)
		if err != nil {
			return nil, err
		}
		if events == nil {
			events = []models.ID{}
		}
		folders[i].EventIds = events
	}
	return folders, nil
}

func GetEventsInFolder(folderID, userID models.ID) ([]models.ID, error) {
	var mappings []models.FolderEvent
	if err := orm.Select("event_id").Where("folder_id = ? AND user_id = ?", folderID, userID).Find(&mappings).Error; err != nil {
		return nil, err
	}
	ids := make([]models.ID, len(mappings))
	for i := range mappings {
		ids[i] = mappings[i].EventId
	}
	return ids, nil
}

func UpdateFolder(folderID, userID models.ID, updates map[string]interface{}) error {
	return orm.Model(&models.Folder{}).Where("id = ? AND user_id = ?", folderID, userID).Updates(updates).Error
}

func SetEventFolder(eventID models.ID, folderID *models.ID, userID models.ID) error {
	return orm.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("event_id = ? AND user_id = ?", eventID, userID).Delete(&models.FolderEvent{}).Error; err != nil {
			return err
		}
		if folderID != nil {
			return tx.Create(&models.FolderEvent{FolderId: *folderID, EventId: eventID, UserId: userID}).Error
		}
		return nil
	})
}

func DeleteFolder(folderID, userID models.ID) error {
	return orm.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Folder{}).Where("id = ? AND user_id = ?", folderID, userID).Update("is_deleted", true).Error; err != nil {
			return err
		}
		var eventIDs []models.ID
		if err := tx.Model(&models.FolderEvent{}).Where("folder_id = ? AND user_id = ?", folderID, userID).Pluck("event_id", &eventIDs).Error; err != nil {
			return err
		}
		if len(eventIDs) > 0 {
			if err := tx.Model(&models.Event{}).Where("id IN ? AND owner_id = ?", eventIDs, userID).Update("is_deleted", true).Error; err != nil {
				return err
			}
		}
		return tx.Where("folder_id = ? AND user_id = ?", folderID, userID).Delete(&models.FolderEvent{}).Error
	})
}
