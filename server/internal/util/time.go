package util

import "time"

// Beijing Timezone (UTC+8)
var CstZone = time.FixedZone("CST", 8*3600)

// Now returns the current time in Beijing Time.
func Now() time.Time {
	return time.Now().In(CstZone)
}
