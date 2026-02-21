package helpers

import (
	"crypto/sha256"
	"encoding/hex"
	"fast-ingest/internal/model"
	"strconv"
)

func NullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func DedupeKey(e model.Event) string {
	ts := e.Timestamp
	if ts > 1e12 {
		ts /= 1000
	} // normalize
	raw := e.EventName + "|" + e.Channel + "|" + e.CampaignID + "|" + e.UserID + "|" + strconv.FormatInt(ts, 10)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
