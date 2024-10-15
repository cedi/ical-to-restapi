package calendar

import (
	"strings"
	"time"

	"github.com/apognu/gocal"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

type EventList struct {
	Date   string  `json:"date"`
	Events []Event `json:"events"`

	zapLog          *otelzap.SugaredLogger
	cacheExpiration time.Time
}

type Event struct {
	Summary string `json:"summary"`
	Start   string `json:"start_date"`
	End     string `json:"end_date"`
	Busy    string `json:"busy"`
	AllDay  bool   `json:"all_day"`
}

func NewEvent(e gocal.Event) *Event {
	if strings.Contains(e.Summary, "Canceled") {
		return nil
	}

	if strings.Contains(e.Summary, "Declined") {
		return nil
	}

	busy := ""
	if val, ok := e.CustomAttributes["X-MICROSOFT-CDO-BUSYSTATUS"]; ok {
		busy = val
	}

	allDay := false
	if val, ok := e.CustomAttributes["X-MICROSOFT-CDO-ALLDAYEVENT"]; ok {
		allDay = val == "TRUE"
	}

	start := e.Start.In(time.Local)
	end := e.End.In(time.Local)

	return &Event{
		Summary: e.Summary,
		Start:   start.Format(time.DateTime),
		End:     end.Format(time.DateTime),
		AllDay:  allDay,
		Busy:    busy,
	}
}
