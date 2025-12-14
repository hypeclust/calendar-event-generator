package templates

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/monil/calendar-event-generator/models"
)

// DateRangeEventInput represents a multi-day event
type DateRangeEventInput struct {
	Name        string   `json:"name"`
	StartDate   string   `json:"start_date"`
	EndDate     string   `json:"end_date"`
	StartTime   string   `json:"start_time,omitempty"` // Optional for non-all-day
	EndTime     string   `json:"end_time,omitempty"`
	AllDay      bool     `json:"all_day,omitempty"`
	Description string   `json:"description,omitempty"`
	Location    string   `json:"location,omitempty"`
	Links       []string `json:"links,omitempty"`
	ColorID     string   `json:"color_id,omitempty"`
}

// DateRangeTemplate represents the date range template format
type DateRangeTemplate struct {
	Format string                `json:"format"`
	Events []DateRangeEventInput `json:"events"`
}

// parseDateRange parses the date range format
func (p *Parser) parseDateRange(data []byte) ([]models.CalendarEvent, error) {
	var template DateRangeTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to parse date range events JSON: %w", err)
	}

	var events []models.CalendarEvent
	for _, dr := range template.Events {
		event, err := p.convertDateRangeEvent(dr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert event '%s': %w", dr.Name, err)
		}
		events = append(events, event)
	}

	return events, nil
}

// convertDateRangeEvent converts a DateRangeEventInput to a CalendarEvent
func (p *Parser) convertDateRangeEvent(dr DateRangeEventInput) (models.CalendarEvent, error) {
	startDate, err := p.TimeParser.ParseDate(dr.StartDate)
	if err != nil {
		return models.CalendarEvent{}, fmt.Errorf("failed to parse start date: %w", err)
	}

	endDate, err := p.TimeParser.ParseDate(dr.EndDate)
	if err != nil {
		return models.CalendarEvent{}, fmt.Errorf("failed to parse end date: %w", err)
	}

	var startTime, endTime time.Time

	// Determine if this is an all-day event
	allDay := dr.AllDay || (dr.StartTime == "" && dr.EndTime == "")

	if allDay {
		// For all-day events in Google Calendar, end date should be exclusive
		startTime = startDate
		endTime = endDate.AddDate(0, 0, 1) // Add one day for exclusive end
	} else {
		// Parse start time
		if dr.StartTime != "" {
			startHour, startMin, err := p.TimeParser.ParseTime(dr.StartTime)
			if err != nil {
				return models.CalendarEvent{}, fmt.Errorf("failed to parse start time: %w", err)
			}
			startTime = p.TimeParser.CombineDateTime(startDate, startHour, startMin)
		} else {
			startTime = startDate // Start of day
		}

		// Parse end time
		if dr.EndTime != "" {
			endHour, endMin, err := p.TimeParser.ParseTime(dr.EndTime)
			if err != nil {
				return models.CalendarEvent{}, fmt.Errorf("failed to parse end time: %w", err)
			}
			endTime = p.TimeParser.CombineDateTime(endDate, endHour, endMin)
		} else {
			// Default to end of the end date
			endTime = p.TimeParser.CombineDateTime(endDate, 23, 59)
		}
	}

	return models.CalendarEvent{
		Name:        dr.Name,
		Description: dr.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		Location:    dr.Location,
		Links:       dr.Links,
		AllDay:      allDay,
		ColorID:     dr.ColorID,
	}, nil
}
