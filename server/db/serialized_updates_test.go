package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"schej.it/server/models"
)

func TestMapUpdatesSerializeEveryJSONField(t *testing.T) {
	database, err := gorm.Open(postgres.Open("host=localhost user=test dbname=test sslmode=disable"), &gorm.Config{
		DryRun:                 true,
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	id := models.NewID()
	calendarOptions := &models.CalendarOptions{}
	signUpBlocks := &[]models.SignUpBlock{{Id: models.NewID(), Name: "block"}}
	remindees := &[]models.Remindee{{Email: "person@example.com", TaskIds: []string{"private-task-id"}}}
	event := &models.Event{
		Id:              models.NewID(),
		Dates:           []models.DateTime{},
		Times:           []models.DateTime{},
		SignUpBlocks:    signUpBlocks,
		SignUpResponses: map[string]*models.SignUpResponse{"guest": {Name: "Guest"}},
		ScheduledEvent:  &models.CalendarEvent{},
		Remindees:       remindees,
	}
	user := &models.User{
		Id: models.NewID(),
		CalendarAccounts: map[string]models.CalendarAccount{
			"person_google": {OAuth2CalendarAuth: &models.OAuth2CalendarAuth{AccessToken: "private-access-token"}},
		},
		CalendarOptions: calendarOptions,
	}
	response := &models.EventResponse{Id: models.NewID(), Response: &models.Response{Name: "Guest"}}
	log := &models.DailyUserLog{Id: models.NewID(), UserIds: []models.ID{id}}

	tests := []struct {
		name        string
		model       interface{}
		updates     map[string]interface{}
		wantJSON    int
		wantContent string
	}{
		{name: "daily user IDs", model: &models.DailyUserLog{}, updates: map[string]interface{}{"user_ids": log.UserIds}, wantJSON: 1},
		{name: "user JSON", model: &models.User{}, updates: map[string]interface{}{
			"calendar_accounts": user.CalendarAccounts,
			"calendar_options":  user.CalendarOptions,
		}, wantJSON: 2, wantContent: "private-access-token"},
		{name: "event JSON", model: &models.Event{}, updates: map[string]interface{}{
			"dates":             event.Dates,
			"times":             event.Times,
			"sign_up_blocks":    event.SignUpBlocks,
			"sign_up_responses": event.SignUpResponses,
			"scheduled_event":   event.ScheduledEvent,
			"remindees":         event.Remindees,
		}, wantJSON: 6, wantContent: "private-task-id"},
		{name: "event response", model: &models.EventResponse{}, updates: map[string]interface{}{"response": response.Response}, wantJSON: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Updates(database.Model(test.model).Where("id = ?", id), test.updates)
			if result.Error != nil {
				t.Fatal(result.Error)
			}

			jsonValues := serializedJSONValues(t, result.Statement.Vars)
			if len(jsonValues) != test.wantJSON {
				t.Fatalf("serialized JSON values = %d, want %d; SQL: %s; vars: %s", len(jsonValues), test.wantJSON, result.Statement.SQL.String(), formatVarTypes(result.Statement.Vars))
			}
			if test.wantContent != "" && !strings.Contains(strings.Join(jsonValues, "\n"), test.wantContent) {
				t.Fatalf("serialized values do not preserve %q: %s", test.wantContent, jsonValues)
			}
		})
	}
}

func TestMapUpdatesPreserveGORMExpressions(t *testing.T) {
	database, err := gorm.Open(postgres.Open("host=localhost user=test dbname=test sslmode=disable"), &gorm.Config{
		DryRun:                 true,
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("clause expression", func(t *testing.T) {
		result := Updates(database.Model(&models.Event{}).Where("id = ?", models.NewID()), map[string]interface{}{
			"remindees": gorm.Expr("COALESCE(remindees, '[]'::jsonb) || ?::jsonb", `[{"email":"person@example.com"}]`),
		})
		if result.Error != nil {
			t.Fatal(result.Error)
		}
		if sql := result.Statement.SQL.String(); !strings.Contains(sql, "COALESCE(remindees") {
			t.Fatalf("expression missing from SQL: %s", sql)
		}
		if got := serializedJSONValues(t, result.Statement.Vars); len(got) != 0 {
			t.Fatalf("expression was serialized: %v", got)
		}
	})

	t.Run("subquery", func(t *testing.T) {
		subquery := database.Table("events AS source").Select("source.remindees").Where("source.id = events.id")
		result := Updates(database.Model(&models.Event{}).Where("id = ?", models.NewID()), map[string]interface{}{
			"remindees": subquery,
		})
		if result.Error != nil {
			t.Fatal(result.Error)
		}
		if sql := result.Statement.SQL.String(); !strings.Contains(sql, "SELECT source.remindees") {
			t.Fatalf("subquery missing from SQL: %s", sql)
		}
		if got := serializedJSONValues(t, result.Statement.Vars); len(got) != 0 {
			t.Fatalf("subquery was serialized: %v", got)
		}
	})
}

func serializedJSONValues(t *testing.T, values []interface{}) []string {
	t.Helper()
	var result []string
	for _, value := range values {
		valuer, ok := value.(driver.Valuer)
		if !ok {
			continue
		}
		serialized, err := valuer.Value()
		if err != nil {
			t.Fatalf("serialize %T: %v", value, err)
		}
		var data []byte
		switch serialized := serialized.(type) {
		case []byte:
			data = serialized
		case string:
			data = []byte(serialized)
		default:
			continue
		}
		if json.Valid(data) {
			result = append(result, string(data))
		}
	}
	return result
}

func formatVarTypes(values []interface{}) string {
	types := make([]string, len(values))
	for index, value := range values {
		types[index] = fmt.Sprintf("%T", value)
	}
	return strings.Join(types, ", ")
}
