package calendar

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/apognu/gocal"
	"github.com/cedi/icaltest/pkg/errors"
)

func (e *EventList) loadEvents(ctx context.Context, from string, url string) ([]Event, *errors.ResolvingError) {
	ical, err := getIcal(ctx, from, url)
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

	events := make([]Event, 0)
	for _, e := range cal.Events {
		event := NewEvent(e)
		if event == nil {
			continue
		}
		events = append(events, *event)
	}

	return events, nil
}
