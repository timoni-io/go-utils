package conv

import (
	"time"
)

// UnixTimeStamp ...
func UnixTimeStamp(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}

	return t.Unix()
}
