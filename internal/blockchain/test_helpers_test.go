package blockchain

import "time"

func mustTime(unix int64) time.Time {
	return time.Unix(unix, 0).UTC()
}
