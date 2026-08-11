package db

import (
	"time"

	"schej.it/server/models"
)

func CountDistinctMonthlyActiveEventCreators(date time.Time) (int64, error) {
	start := date.AddDate(0, 0, -30)
	var count int64
	err := orm.Model(&models.Event{}).
		Where("id >= ? AND id < ?", models.NewIDFromTimestamp(start), models.NewIDFromTimestamp(date)).
		Where("creator_posthog_id IS NOT NULL AND creator_posthog_id <> ''").
		Distinct("creator_posthog_id").Count(&count).Error
	return count, err
}

func CountDistinctMonthlyActiveEventCreatorsWithMoreThanXEvents(date time.Time, minimum int) (int64, error) {
	start := date.AddDate(0, 0, -30)
	var count int64
	subquery := orm.Model(&models.Event{}).
		Select("creator_posthog_id").
		Where("id >= ? AND id < ?", models.NewIDFromTimestamp(start), models.NewIDFromTimestamp(date)).
		Where("creator_posthog_id IS NOT NULL AND creator_posthog_id <> ''").
		Group("creator_posthog_id").Having("COUNT(*) >= ?", minimum)
	err := orm.Table("(?) AS active_creators", subquery).Count(&count).Error
	return count, err
}
