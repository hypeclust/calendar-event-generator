package templates

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/monil/calendar-event-generator/models"
)

// WeeklyEvent represents an event in the weekly schedule format
type WeeklyEvent struct {
	EventName    string   `json:"event_name"`
	Date         string   `json:"date"`
	Time         string   `json:"time"`
	TopicDetails string   `json:"topic_details"`
	UsefulLinks  []string `json:"useful_links"`
	Location     string   `json:"location,omitempty"`
	Description  string   `json:"description,omitempty"`
}

// parseWeekly parses the weekly schedule format
// Format: { "week_1": [...], "week_2": [...] }
func (p *Parser) parseWeekly(data []byte) ([]models.CalendarEvent, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse weekly JSON: %w", err)
	}

	var events []models.CalendarEvent

	// Sort keys to process weeks in order
	var keys []string
	for key := range raw {
		// Skip non-week keys like "format" or "metadata"
		if strings.HasPrefix(strings.ToLower(key), "week") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	for _, weekKey := range keys {
		var weekEvents []WeeklyEvent
		if err := json.Unmarshal(raw[weekKey], &weekEvents); err != nil {
			return nil, fmt.Errorf("failed to parse week %s: %w", weekKey, err)
		}

		for _, we := range weekEvents {
			event, err := p.convertWeeklyEvent(we)
			if err != nil {
				return nil, fmt.Errorf("failed to convert event '%s': %w", we.EventName, err)
			}
			events = append(events, event)
		}
	}

	return events, nil
}

// convertWeeklyEvent converts a WeeklyEvent to a CalendarEvent
func (p *Parser) convertWeeklyEvent(we WeeklyEvent) (models.CalendarEvent, error) {
	// Parse date and time range
	startTime, endTime, err := p.TimeParser.ParseDateTimeRange(we.Date, we.Time)
	if err != nil {
		return models.CalendarEvent{}, fmt.Errorf("failed to parse date/time: %w", err)
	}

	// Build description from topic details
	description := we.TopicDetails
	if we.Description != "" {
		if description != "" {
			description += "\n\n"
		}
		description += we.Description
	}

	return models.CalendarEvent{
		Name:        we.EventName,
		Description: description,
		StartTime:   startTime,
		EndTime:     endTime,
		Location:    we.Location,
		Links:       we.UsefulLinks,
	}, nil
}
