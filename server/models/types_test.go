package models

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm/schema"
)

func TestIDRoundTrip(t *testing.T) {
	id := NewID()
	parsed, err := ParseID(id.Hex())
	if err != nil {
		t.Fatalf("ParseID() error = %v", err)
	}
	if parsed != id {
		t.Fatalf("ParseID() = %q, want %q", parsed, id)
	}
	if delta := time.Since(id.Timestamp()); delta < 0 || delta > 2*time.Second {
		t.Fatalf("ID timestamp delta = %v", delta)
	}
}

func TestParseIDCanonicalizesUppercase(t *testing.T) {
	id := NewID()
	parsed, err := ParseID(strings.ToUpper(id.Hex()))
	if err != nil {
		t.Fatalf("ParseID() error = %v", err)
	}
	if parsed != id {
		t.Fatalf("ParseID() = %q, want %q", parsed, id)
	}
}

func TestStorageJSONKeepsPrivateFields(t *testing.T) {
	accounts := map[string]CalendarAccount{
		"person_google": {
			OAuth2CalendarAuth: &OAuth2CalendarAuth{AccessToken: "access", RefreshToken: "refresh"},
		},
	}
	data, err := marshalStorageJSON(accounts)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"accessToken":"access"`) || !strings.Contains(string(data), `"refreshToken":"refresh"`) {
		t.Fatalf("stored credentials missing: %s", data)
	}

	var restored map[string]CalendarAccount
	if err := unmarshalStorageJSON(data, reflect.ValueOf(&restored).Elem()); err != nil {
		t.Fatal(err)
	}
	if restored["person_google"].OAuth2CalendarAuth.AccessToken != "access" {
		t.Fatalf("stored credential not restored: %#v", restored)
	}

	remindees, err := marshalStorageJSON([]Remindee{{Email: "x@example.com", TaskIds: []string{"task"}}})
	if err != nil || !strings.Contains(string(remindees), `"taskIds":["task"]`) {
		t.Fatalf("stored task IDs missing: %s, %v", remindees, err)
	}
}

func TestStorageJSONUsesBSONTags(t *testing.T) {
	data, err := marshalStorageJSON(Location{CountryCode: "US"})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != `{"countryCode":"US"}` {
		t.Fatalf("storage JSON = %s, want BSON field names with empty values omitted", got)
	}

	var restored Location
	if err := unmarshalStorageJSON(data, reflect.ValueOf(&restored).Elem()); err != nil {
		t.Fatal(err)
	}
	if restored.CountryCode != "US" {
		t.Fatalf("country code = %q, want US", restored.CountryCode)
	}
}

func TestDateTimeJSONRoundTrip(t *testing.T) {
	want := time.Date(2026, time.August, 9, 12, 30, 0, 0, time.UTC)
	data, err := json.Marshal(NewDateTime(want))
	if err != nil {
		t.Fatal(err)
	}
	var got DateTime
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if !got.Time().Equal(want) {
		t.Fatalf("DateTime = %v, want %v", got.Time(), want)
	}
}

func TestGormModelsParse(t *testing.T) {
	models := []interface{}{&User{}, &Event{}, &EventResponse{}, &Attendee{}, &Folder{}, &FolderEvent{}, &DailyUserLog{}, &FriendRequest{}, &OtpCode{}}
	cache := &sync.Map{}
	for _, model := range models {
		if _, err := schema.Parse(model, cache, schema.NamingStrategy{}); err != nil {
			t.Errorf("schema.Parse(%T): %v", model, err)
		}
	}
}
