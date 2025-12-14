package calendar

import (
	"fmt"
	"time"

	"github.com/monil/calendar-event-generator/models"
	"google.golang.org/api/calendar/v3"
)

// EventResult represents the result of creating an event
type EventResult struct {
	Event   *models.CalendarEvent
	GEvent  *calendar.Event
	Success bool
	Error   error
	Link    string
}

// CreateEvent creates a single event in Google Calendar
func (c *Client) CreateEvent(event *models.CalendarEvent) (*EventResult, error) {
	gEvent := c.convertToGoogleEvent(event)

	created, err := c.service.Events.Insert(c.calendarID, gEvent).Do()
	if err != nil {
		return &EventResult{
			Event:   event,
			Success: false,
			Error:   err,
		}, err
	}

	return &EventResult{
		Event:   event,
		GEvent:  created,
		Success: true,
		Link:    created.HtmlLink,
	}, nil
}

// CreateEvents creates multiple events with progress reporting
func (c *Client) CreateEvents(events []models.CalendarEvent, callback func(int, int, *EventResult)) ([]*EventResult, error) {
	results := make([]*EventResult, len(events))

	for i, event := range events {
		result, _ := c.CreateEvent(&event)
		results[i] = result

		if callback != nil {
			callback(i+1, len(events), result)
		}

		// Small delay to avoid rate limiting
		if i < len(events)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return results, nil
}

// convertToGoogleEvent converts a CalendarEvent to a Google Calendar Event
func (c *Client) convertToGoogleEvent(event *models.CalendarEvent) *calendar.Event {
	gEvent := &calendar.Event{
		Summary:     event.Name,
		Description: event.FormatDescription(),
		Location:    event.Location,
	}

	// Set start and end times
	if event.AllDay {
		gEvent.Start = &calendar.EventDateTime{
			Date: event.StartTime.Format("2006-01-02"),
		}
		gEvent.End = &calendar.EventDateTime{
			Date: event.EndTime.Format("2006-01-02"),
		}
	} else {
		gEvent.Start = &calendar.EventDateTime{
			DateTime: event.StartTime.Format(time.RFC3339),
		}
		if tz := event.StartTime.Location().String(); tz != "Local" {
			gEvent.Start.TimeZone = tz
		}

		gEvent.End = &calendar.EventDateTime{
			DateTime: event.EndTime.Format(time.RFC3339),
		}
		if tz := event.EndTime.Location().String(); tz != "Local" {
			gEvent.End.TimeZone = tz
		}
	}

	// Set recurrence rule
	if event.Recurrence != nil {
		rrule := c.buildRRule(event.Recurrence)
		if rrule != "" {
			gEvent.Recurrence = []string{rrule}
		}
	}

	// Set color
	if event.ColorID != "" {
		gEvent.ColorId = event.ColorID
	}

	// Set reminders
	if len(event.Reminders) > 0 {
		overrides := make([]*calendar.EventReminder, len(event.Reminders))
		for i, r := range event.Reminders {
			overrides[i] = &calendar.EventReminder{
				Method:  r.Method,
				Minutes: int64(r.Minutes),
			}
		}
		gEvent.Reminders = &calendar.EventReminders{
			UseDefault: false,
			Overrides:  overrides,
		}
	}

	return gEvent
}

// buildRRule creates an RRULE string from RecurrenceRule
func (c *Client) buildRRule(r *models.RecurrenceRule) string {
	if r == nil {
		return ""
	}

	rule := "RRULE:FREQ=" + r.Frequency

	if r.Interval > 1 {
		rule += fmt.Sprintf(";INTERVAL=%d", r.Interval)
	}

	if r.Until != nil {
		rule += ";UNTIL=" + r.Until.UTC().Format("20060102T150405Z")
	}

	if r.Count > 0 {
		rule += fmt.Sprintf(";COUNT=%d", r.Count)
	}

	if len(r.ByDay) > 0 {
		rule += ";BYDAY="
		for i, day := range r.ByDay {
			if i > 0 {
				rule += ","
			}
			rule += day
		}
	}

	return rule
}

// DryRunEvent validates an event without creating it
func (c *Client) DryRunEvent(event *models.CalendarEvent) *calendar.Event {
	return c.convertToGoogleEvent(event)
}

// DryRunEvents validates multiple events and returns Google Calendar event representations
func (c *Client) DryRunEvents(events []models.CalendarEvent) []*calendar.Event {
	gEvents := make([]*calendar.Event, len(events))
	for i, e := range events {
		gEvents[i] = c.DryRunEvent(&e)
	}
	return gEvents
}
