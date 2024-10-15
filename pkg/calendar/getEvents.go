package calendar

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func (e *EventList) CacheInvalidate() {
	e.cacheExpiration = time.Now().Add(-1 * time.Hour)
}

func (eventList *EventList) GetEvents(ct *gin.Context) {
	now := time.Now().Format(time.DateOnly)

	if time.Now().Before(eventList.cacheExpiration) {
		ct.JSON(http.StatusOK, eventList)
		return
	}

	var eventsMux sync.Mutex
	eventList.Date = now
	eventList.Events = make([]Event, 0)

	calendars := viper.GetStringMap("calendars")

	var wg sync.WaitGroup

	for key := range calendars {
		from := viper.GetString(fmt.Sprintf("calendars.%s.from", key))
		url := viper.GetString(fmt.Sprintf("calendars.%s.ical", key))
		wg.Add(1)
		go func() {
			ctx := ct.Request.Context()
			events, err := eventList.loadEvents(ctx, from, url)
			if err != nil {
				eventList.zapLog.Ctx(ctx).Errorw("Unable to load events", "error", err.Err, "how_to_fix", err.HowToResolve)
			}

			eventsMux.Lock()
			eventList.Events = append(eventList.Events, events...)
			eventsMux.Unlock()

			wg.Done()
		}()
	}
	wg.Wait()

	eventList.cacheExpiration = time.Now().Add(30 * time.Minute)

	ct.JSON(http.StatusOK, eventList)
}
