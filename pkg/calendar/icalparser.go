package calendar

import (
	"fmt"
	"time"

	"github.com/apognu/gocal"
)

var tzMapping = map[string]string{
	"Romance Standard Time":        "Europe/Brussels",
	"Pacific Standard Time":        "US/Pacific",
	"W. Europe Standard Time":      "Europe/Berlin",
	"E. Australia Standard Time":   "Australia/Brisbane",
	"GMT Standard Time":            "Europe/Dublin",
	"Eastern Standard Time":        "US/Eastern",
	"Greenwich Standard Time":      "Etc/GMT",
	"\tzone://Microsoft/Utc\"":     "UTC",
	"Central Europe Standard Time": "Europe/Berlin",
	"Central Standard Time":        "US/Central",
	"Customized Time Zone":         "UTC",
	"India Standard Time":          "Asia/Calcutta",
	"AUS Eastern Standard Time":    "Australia/Brisbane",
	"UTC":                          "UTC",
	"Israel Standard Time":         "Israel",
	"Singapore Standard Time":      "Singapore",
}

func init() {
	gocal.SetTZMapper(func(s string) (*time.Location, error) {
		if tzid, ok := tzMapping[s]; ok {
			return time.LoadLocation(tzid)
		}
		return nil, fmt.Errorf("")
	})
}
