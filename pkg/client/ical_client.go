package client

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/SpechtLabs/CalendarAPI/pkg/errors"
	"github.com/apognu/gocal"
	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"

	pb "github.com/SpechtLabs/CalendarAPI/pkg/protos"
)

type ICalClient struct {
	cacheMux        sync.RWMutex
	cache           *pb.CalendarResponse
	cacheExpiration time.Time
	zapLog          *otelzap.Logger

	statusMux    sync.RWMutex
	CustomStatus map[string]*pb.CustomStatus // custom status is a map from calendar-name to status
}

type Calendar struct {
	Name string `mapstructure:"name"`
	From string `mapstructure:"from"`
	Ical string `mapstructure:"ical"`
}

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

func parseCalendars() []Calendar {
	calendars := []Calendar{}
	viper.UnmarshalKey("calendars", &calendars)
	return calendars
}

func NewICalClient(zapLog *otelzap.Logger) *ICalClient {
	return &ICalClient{
		zapLog:          zapLog,
		cacheExpiration: time.Now(),
		cache:           &pb.CalendarResponse{LastUpdated: time.Now().Unix()},
		CustomStatus:    make(map[string]*pb.CustomStatus),
	}
}

func (e *ICalClient) FetchEvents(ctx context.Context) {
	response := &pb.CalendarResponse{
		LastUpdated: time.Now().Unix(),
		Entries:     make([]*pb.CalendarEntry, 0),
	}

	calendars := parseCalendars()
	rules := parseRules()

	var wg sync.WaitGroup
	var eventsMux sync.Mutex

	for _, cal := range calendars {
		name := cal.Name
		from := cal.From
		url := cal.Ical

		wg.Add(1)

		go func() {
			start := time.Now()
			events, err := e.loadEvents(ctx, name, from, url, rules)
			stop := time.Now()
			if err != nil {
				e.zapLog.Ctx(ctx).Sugar().Errorw("Unable to load events", err.AsZapLogKV())
			}

			eventsMux.Lock()
			response.LastUpdated = time.Now().Unix()
			response.Entries = append(response.Entries, events...)
			eventsMux.Unlock()

			e.zapLog.Ctx(ctx).Sugar().Infof("Refreshed calendar %s in %dms", name, stop.Sub(start).Milliseconds())

			wg.Done()
		}()
	}

	wg.Wait()
	e.cacheMux.Lock()
	e.cache = response
	e.cacheMux.Unlock()
}

func (e *ICalClient) GetEvents(ctx context.Context) *pb.CalendarResponse {
	if e.cache == nil {
		e.zapLog.Ctx(ctx).Sugar().Infow("Experiencing cold. Fetching events now!")
		e.FetchEvents(ctx)
	}

	e.cacheMux.RLock()
	defer e.cacheMux.RUnlock()
	return e.cache
}

func (e *ICalClient) GetCurrentEvent(ctx context.Context, calendar string) *pb.CalendarEntry {
	if e.cache == nil {
		e.zapLog.Ctx(ctx).Sugar().Infow("Experiencing cold. Fetching events now!")
		e.FetchEvents(ctx)
	}

	e.cacheMux.RLock()
	defer e.cacheMux.RUnlock()

	var possibleCurrentEvents []*pb.CalendarEntry

	// Find all events happening right now
	now := time.Now().Unix()
	for _, entry := range e.cache.Entries {
		if calendar != "all" && entry.CalendarName != calendar {
			continue
		}

		if entry.Start < now && entry.End > now {
			possibleCurrentEvents = append(possibleCurrentEvents, entry)
		}
	}

	// If no events or only one event, return early
	switch len(possibleCurrentEvents) {
	case 0:
		return nil
	case 1:
		return possibleCurrentEvents[0]
	}

	// Find the event that starts or ends closest to now
	var closest *pb.CalendarEntry
	closestDelta := int64(math.MaxInt64)

	for _, entry := range possibleCurrentEvents {
		delta := int64(now) - int64(entry.Start)

		if delta == closestDelta && entry.Important && (closest == nil || !closest.Important) {
			closest = entry
		} else if delta < closestDelta {
			closest = entry
			closestDelta = delta
		}
	}

	return closest
}

func (e *ICalClient) GetCustomStatus(ctx context.Context, req *pb.GetCustomStatusRequest) *pb.CustomStatus {
	e.statusMux.RLock()
	defer e.statusMux.RUnlock()

	if val, ok := e.CustomStatus[req.CalendarName]; ok {
		return val
	}

	return &pb.CustomStatus{}
}

func (e *ICalClient) SetCustomStatus(ctx context.Context, req *pb.SetCustomStatusRequest) {
	e.statusMux.Lock()
	defer e.statusMux.Unlock()

	e.CustomStatus[req.CalendarName] = req.Status
}

func (e *ICalClient) loadEvents(ctx context.Context, calName string, from string, url string, rules []Rule) ([]*pb.CalendarEntry, *errors.ResolvingError) {
	ical, err := e.getIcal(ctx, from, url)
	if ical == nil || err != nil {
		return nil, errors.Wrap(err, fmt.Errorf("failed to load iCal calendar file"), "")
	}

	defer ical.Close()
	cal := gocal.NewParser(ical)

	// Filter to TODAY only
	today, _ := time.Parse(time.DateOnly, time.Now().Format(time.DateOnly))
	eod := today.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	start, end := today, eod
	cal.Start, cal.End = &start, &end

	if err := cal.Parse(); err != nil {
		return nil, errors.NewResolvingError(fmt.Errorf("unable to parse iCal file %w", err), "ensure the iCal file is valid and follows the iCal spec")
	}

	// Sort Events by start-date (makes our live easier down the line)
	sort.Slice(cal.Events, func(i int, j int) bool {
		left := cal.Events[i]
		right := cal.Events[j]
		return left.Start.Before(*right.Start)
	})

	events := make([]*pb.CalendarEntry, 0)
	for _, evnt := range cal.Events {
		event := NewCalendarEntryFromGocalEvent(calName, evnt)
		if event == nil {
			continue
		}

		// let's evaluate our rules
		for _, rule := range rules {
			// if a rule is sucessfully evaluated
			if ok, skip := rule.Evaluate(event, e.zapLog); ok {
				// if this is a skip rule, don't process any other rules for this
				// event and don't add it
				if skip {
					break
				}

				events = append(events, event)

				// since we found the first rule that matches, no need to
				// process any more rules
				break
			}
		}
	}

	return events, nil
}

func NewCalendarEntryFromGocalEvent(calName string, e gocal.Event) *pb.CalendarEntry {
	if strings.Contains(e.Summary, "Canceled") {
		return nil
	}

	if strings.Contains(e.Summary, "Declined") {
		return nil
	}

	busy := pb.BusyState_Free
	if val, ok := e.CustomAttributes["X-MICROSOFT-CDO-BUSYSTATUS"]; ok {
		switch val {
		case "BUSY":
			busy = pb.BusyState_Busy
		case "TENTATIVE":
			busy = pb.BusyState_Tentative

		case "FREE":
			busy = pb.BusyState_Free

		case "OOF":
			busy = pb.BusyState_OutOfOffice

		case "WORKINGELSEWHERE":
			busy = pb.BusyState_WorkingElsewhere
		}
	}

	allDay := false
	if val, ok := e.CustomAttributes["X-MICROSOFT-CDO-ALLDAYEVENT"]; ok {
		allDay = val == "TRUE"
	}

	start := e.Start.In(time.Local)
	end := e.End.In(time.Local)

	return &pb.CalendarEntry{
		Title:        e.Summary,
		Start:        start.Unix(),
		End:          end.Unix(),
		AllDay:       allDay,
		Busy:         busy,
		CalendarName: calName,
	}
}

func (e *ICalClient) getIcal(ctx context.Context, from string, url string) (io.ReadCloser, *errors.ResolvingError) {
	switch from {
	case "file":
		return e.getIcalFromFile(url)
	case "url":
		return e.getIcalFromURL(ctx, url)
	default:
		return nil, errors.NewResolvingError(fmt.Errorf("unsupported 'from' type"), "The only supported values for 'from' are 'file' or 'url'")
	}
}

func (e *ICalClient) getIcalFromFile(path string) (io.ReadCloser, *errors.ResolvingError) {
	file, err := os.Open(path)
	return file, errors.NewResolvingError(err, "check if file path exists and is accessible")
}

func (e *ICalClient) getIcalFromURL(ctx context.Context, url string) (io.ReadCloser, *errors.ResolvingError) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.NewResolvingError(fmt.Errorf("failed creating request for %s: %w", url, err), "")
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.NewResolvingError(fmt.Errorf("failed making request to %s: %w", url, err), "verify if URL exists and is accessible")
	}

	return resp.Body, nil
}
