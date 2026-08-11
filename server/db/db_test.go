package db_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"schej.it/server/db"
	"schej.it/server/models"
)

func requirePostgresIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("RUN_POSTGRES_INTEGRATION_TESTS") != "1" {
		t.Skip("set RUN_POSTGRES_INTEGRATION_TESTS=1 to run PostgreSQL integration tests")
	}
	closeConnection := db.Init()
	t.Cleanup(closeConnection)
}

func TestGetDailyUserLogByDate(t *testing.T) {
	requirePostgresIntegration(t)
	db.GetDailyUserLogByDate(time.Now(), 7)
}

func TestGenerateShortEventId(t *testing.T) {
	requirePostgresIntegration(t)

	eventLongID, _ := models.ParseID("6607d6409f96021811c0a55f")
	id := db.GenerateShortEventId(eventLongID)
	fmt.Println(id)
}
