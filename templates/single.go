package templates

import (
	"encoding/json"
	"fmt"

	"github.com/monil/calendar-event-generator/models"
)

// SingleEventInput represents a single event in the simple format
type SingleEventInput struct {
	Name        string   `json:"name"`
	Date        string   `json:"date"`
	StartTime   string   `json:"start_time"`
	EndTime     string   `json:"end_time"`
	Duration    string   `json:"duration,omitempty"` // Alternative to end_time
	Description string   `json:"description,omitempty"`
	Location    string   `json:"location,omitempty"`
	Links       []string `json:"links,omitempty"`
	AllDay      bool     `json:"all_day,omitempty"`
	ColorID     string   `json:"color_id,omitempty"`
}

// SingleTemplate represents the single events template format
type SingleTemplate struct {
	Format string             `json:"format"`
	Events []SingleEventInput `json:"events"`
}

// parseSingle parses the single event format
func (p *Parser) parseSingle(data []byte) ([]models.CalendarEvent, error) {
	var template SingleTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to parse single events JSON: %w", err)
	}

	var events []models.CalendarEvent
	for _, se := range template.Events {
		event, err := p.convertSingleEvent(se)
		if err != nil {
			return nil, fmt.Errorf("failed to convert event '%s': %w", se.Name, err)
		}
		events = append(events, event)
	}

	return events, nil
}

// convertSingleEvent converts a SingleEventInput to a CalendarEvent
func (p *Parser) convertSingleEvent(se SingleEventInput) (models.CalendarEvent, error) {
	date, err := p.TimeParser.ParseDate(se.Date)
	if err != nil {
		return models.CalendarEvent{}, fmt.Errorf("failed to parse date: %w", err)
	}

	var startTime, endTime = date, date

	if !se.AllDay {
		// Parse start time
		startHour, startMin, err := p.TimeParser.ParseTime(se.StartTime)
		if err != nil {
			return models.CalendarEvent{}, fmt.Errorf("failed to parse start time: %w", err)
		}
		startTime = p.TimeParser.CombineDateTime(date, startHour, startMin)

		// Parse end time or calculate from duration
		if se.EndTime != "" {
			endHour, endMin, err := p.TimeParser.ParseTime(se.EndTime)
			if err != nil {
				return models.CalendarEvent{}, fmt.Errorf("failed to parse end time: %w", err)
			}
			endTime = p.TimeParser.CombineDateTime(date, endHour, endMin)

			// Handle overnight events
			if endTime.Before(startTime) {
				endTime = endTime.AddDate(0, 0, 1)
			}
		} else if se.Duration != "" {
			duration, err := p.TimeParser.ParseDuration(se.Duration)
			if err != nil {
				return models.CalendarEvent{}, fmt.Errorf("failed to parse duration: %w", err)
			}
			endTime = startTime.Add(duration)
		} else {
			// Default 1 hour duration
			endTime = startTime.Add(1 * 60 * 60 * 1e9) // 1 hour in nanoseconds
		}
	} else {
		// All-day event - set to start and end of day
		endTime = date.AddDate(0, 0, 1)
	}

	return models.CalendarEvent{
		Name:        se.Name,
		Description: se.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		Location:    se.Location,
		Links:       se.Links,
		AllDay:      se.AllDay,
		ColorID:     se.ColorID,
	}, nil
}
