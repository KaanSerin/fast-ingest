package storage

import (
	"fast-ingest/internal/model"
	"testing"
	"time"
)

func TestNullIfEmpty(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNil bool
		wantVal string
	}{
		{"empty string returns nil", "", true, ""},
		{"non-empty string returns value", "web", false, "web"},
		{"whitespace is not empty", " ", false, " "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NullIfEmpty(tt.input)
			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("expected %q, got nil", tt.wantVal)
				} else if result.(string) != tt.wantVal {
					t.Errorf("expected %q, got %v", tt.wantVal, result)
				}
			}
		})
	}
}

func TestDedupeKey(t *testing.T) {
	base := model.Event{
		EventName:  "page_view",
		Channel:    "web",
		CampaignID: "camp_1",
		UserID:     "user_123",
		Timestamp:  1769904000,
	}

	t.Run("same event produces same key", func(t *testing.T) {
		key1 := DedupeKey(base)
		key2 := DedupeKey(base)
		if key1 != key2 {
			t.Errorf("expected same key, got %q and %q", key1, key2)
		}
	})

	t.Run("returns 64-char hex string", func(t *testing.T) {
		key := DedupeKey(base)
		if len(key) != 64 {
			t.Errorf("expected 64 chars, got %d: %q", len(key), key)
		}
	})

	t.Run("different event_name produces different key", func(t *testing.T) {
		other := base
		other.EventName = "click"
		if DedupeKey(base) == DedupeKey(other) {
			t.Error("expected different keys for different event_name")
		}
	})

	t.Run("different user_id produces different key", func(t *testing.T) {
		other := base
		other.UserID = "user_456"
		if DedupeKey(base) == DedupeKey(other) {
			t.Error("expected different keys for different user_id")
		}
	})

	t.Run("different channel produces different key", func(t *testing.T) {
		other := base
		other.Channel = "mobile"
		if DedupeKey(base) == DedupeKey(other) {
			t.Error("expected different keys for different channel")
		}
	})

	t.Run("different timestamp produces different key", func(t *testing.T) {
		other := base
		other.Timestamp = 1769904001
		if DedupeKey(base) == DedupeKey(other) {
			t.Error("expected different keys for different timestamp")
		}
	})

	t.Run("millisecond timestamp normalized to same key as second timestamp", func(t *testing.T) {
		msEvent := base
		msEvent.Timestamp = base.Timestamp * 1000
		if DedupeKey(base) != DedupeKey(msEvent) {
			t.Errorf("expected same key after ms normalization: %q vs %q", DedupeKey(base), DedupeKey(msEvent))
		}
	})
}

func TestNormalizeTimestamp(t *testing.T) {
	t.Run("second timestamp returns correct UTC time", func(t *testing.T) {
		ts := int64(1769904000)
		result := NormalizeTimestamp(ts)
		expected := time.Unix(1769904000, 0).UTC()
		if !result.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})

	t.Run("millisecond timestamp normalizes to same time as second timestamp", func(t *testing.T) {
		ts := int64(1769904000)
		resultSec := NormalizeTimestamp(ts)
		resultMs := NormalizeTimestamp(ts * 1000)
		if !resultSec.Equal(resultMs) {
			t.Errorf("expected %v, got %v", resultSec, resultMs)
		}
	})

	t.Run("result is always UTC", func(t *testing.T) {
		result := NormalizeTimestamp(1769904000)
		if result.Location() != time.UTC {
			t.Errorf("expected UTC, got %v", result.Location())
		}
	})

	t.Run("boundary: exactly 1e12 is treated as seconds", func(t *testing.T) {
		ts := int64(1e12)
		result := NormalizeTimestamp(ts)
		expected := time.Unix(ts, 0).UTC()
		if !result.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})

	t.Run("boundary: 1e12+1 is treated as milliseconds", func(t *testing.T) {
		ts := int64(1e12) + 1
		result := NormalizeTimestamp(ts)
		expected := time.Unix(ts/1000, 0).UTC()
		if !result.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})
}
