package models

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ID preserves the public 24-hex-character identifier format while remaining
// independent from any database driver.
type ID string

const NilID ID = ""

func NewID() ID {
	var raw [12]byte
	seconds := uint32(time.Now().Unix())
	raw[0], raw[1], raw[2], raw[3] = byte(seconds>>24), byte(seconds>>16), byte(seconds>>8), byte(seconds)
	if _, err := rand.Read(raw[4:]); err != nil {
		panic(fmt.Errorf("generate ID: %w", err))
	}
	return ID(hex.EncodeToString(raw[:]))
}

func NewIDFromTimestamp(t time.Time) ID {
	var raw [12]byte
	seconds := uint32(t.Unix())
	raw[0], raw[1], raw[2], raw[3] = byte(seconds>>24), byte(seconds>>16), byte(seconds>>8), byte(seconds)
	return ID(hex.EncodeToString(raw[:]))
}

func ParseID(value string) (ID, error) {
	if len(value) != 24 {
		return NilID, errors.New("ID must contain 24 hexadecimal characters")
	}
	if _, err := hex.DecodeString(value); err != nil {
		return NilID, fmt.Errorf("invalid ID: %w", err)
	}
	return ID(strings.ToLower(value)), nil
}

func (id ID) Hex() string  { return string(id) }
func (id ID) IsZero() bool { return id == NilID }
func (id ID) Timestamp() time.Time {
	if len(id) < 8 {
		return time.Time{}
	}
	seconds, err := strconv.ParseUint(string(id[:8]), 16, 32)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(int64(seconds), 0)
}
func (id ID) Value() (driver.Value, error) { return string(id), nil }
func (id *ID) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		*id = ID(v)
	case []byte:
		*id = ID(v)
	case nil:
		*id = NilID
	default:
		return fmt.Errorf("scan ID from %T", value)
	}
	return nil
}

// DateTime stores Unix milliseconds, matching existing API semantics.
type DateTime int64

func NewDateTime(t time.Time) DateTime          { return DateTime(t.UnixMilli()) }
func (d DateTime) Time() time.Time              { return time.UnixMilli(int64(d)) }
func (d DateTime) Value() (driver.Value, error) { return d.Time(), nil }
func (d *DateTime) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		*d = NewDateTime(v)
	case nil:
		*d = 0
	default:
		return fmt.Errorf("scan DateTime from %T", value)
	}
	return nil
}
func (d DateTime) MarshalJSON() ([]byte, error) { return json.Marshal(d.Time()) }
func (d *DateTime) UnmarshalJSON(data []byte) error {
	var t time.Time
	if err := json.Unmarshal(data, &t); err == nil {
		*d = NewDateTime(t)
		return nil
	}
	var millis int64
	if err := json.Unmarshal(data, &millis); err != nil {
		return err
	}
	*d = DateTime(millis)
	return nil
}
