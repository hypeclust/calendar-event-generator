package exporter

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	ical "github.com/arran4/golang-ical"
	"github.com/monil/calendar-event-generator/models"
)

// GenerateICS converts a list of CalendarEvents to an iCalendar file content
func GenerateICS(events []models.CalendarEvent, w io.Writer) error {
	cal := ical.NewCalendar()
	cal.SetMethod(ical.MethodRequest)
	cal.SetProductId("-//Monil//Calendar Event Generator//EN")
	cal.SetVersion("2.0")

	for _, e := range events {
		event := cal.AddEvent(generateUID(e))
		event.SetSummary(e.Name)
		
		if e.Description != "" {
			desc := e.FormatDescription()
			event.SetDescription(desc)
		}

		if e.Location != "" {
			event.SetLocation(e.Location)
		}

		event.SetDtStampTime(time.Now())

		if e.AllDay {
			// All day events require standard date format (YYYYMMDD)
			event.SetProperty(ical.ComponentPropertyDtStart, e.StartTime.Format("20060102"), ical.WithValue("DATE"))
			// End date for all day events is exclusive, so add 1 day if not set or same as start
			endTime := e.EndTime
			if endTime.IsZero() || endTime.Equal(e.StartTime) {
				endTime = e.StartTime.AddDate(0, 0, 1)
			} else {
				// If end date is present, ensure it's treated as next day for exclusive end
				endTime = endTime.AddDate(0, 0, 1)
			}
			event.SetProperty(ical.ComponentPropertyDtEnd, endTime.Format("20060102"), ical.WithValue("DATE"))
		} else {
			event.SetStartAt(e.StartTime)
			if !e.EndTime.IsZero() {
				event.SetEndAt(e.EndTime)
			} else {
				// Default 1 hour if no end time? Or just start?
				// Let's default to start + 1h if missing
				event.SetEndAt(e.StartTime.Add(time.Hour))
			}
		}

		if e.Recurrence != nil {
			rrule := e.Recurrence.ToRRuleString()
			// Remove "RRULE:" prefix as library might add it or we set property directly
			// golang-ical SetProperty takes value. RRuleString includes key.
			// Let's assume we pass value.
			val := strings.TrimPrefix(rrule, "RRULE:")
			event.SetProperty(ical.ComponentPropertyRrule, val)
		}
	}

	return cal.SerializeTo(w)
}

func generateUID(e models.CalendarEvent) string {
	// Simple deterministic UID based on content
	data := fmt.Sprintf("%s-%s-%s", e.Name, e.StartTime.String(), e.Description)
	hash := sha1.Sum([]byte(data))
	return hex.EncodeToString(hash[:]) + "@calendar-generator"
}
