package templates

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/monil/calendar-event-generator/models"
)

// RecurrenceInput represents recurrence settings in the JSON
type RecurrenceInput struct {
	Frequency       string   `json:"frequency"` // DAILY, WEEKLY, MONTHLY, YEARLY
	Interval        int      `json:"interval,omitempty"`
	Until           string   `json:"until,omitempty"`
	Count           int      `json:"count,omitempty"`
	ByDay           []string `json:"by_day,omitempty"`
	ExcludeWeekends bool     `json:"exclude_weekends,omitempty"`
}

// RecurringEventInput represents a recurring event in the JSON
type RecurringEventInput struct {
	Name        string          `json:"name"`
	StartDate   string          `json:"start_date,omitempty"` // Optional start date
	StartTime   string          `json:"start_time"`
	EndTime     string          `json:"end_time,omitempty"`
	Duration    string          `json:"duration,omitempty"`
	Description string          `json:"description,omitempty"`
	Location    string          `json:"location,omitempty"`
	Links       []string        `json:"links,omitempty"`
	Recurrence  RecurrenceInput `json:"recurrence"`
	ColorID     string          `json:"color_id,omitempty"`
}

// RecurringTemplate represents the recurring events template format
type RecurringTemplate struct {
	Format string                `json:"format"`
	Events []RecurringEventInput `json:"events"`
}

// parseRecurring parses the recurring event format
func (p *Parser) parseRecurring(data []byte) ([]models.CalendarEvent, error) {
	var template RecurringTemplate
	if err := json.Unmarshal(data, &template); err != nil {
		return nil, fmt.Errorf("failed to parse recurring events JSON: %w", err)
	}

	var events []models.CalendarEvent
	for _, re := range template.Events {
		event, err := p.convertRecurringEvent(re)
		if err != nil {
			return nil, fmt.Errorf("failed to convert event '%s': %w", re.Name, err)
		}
		events = append(events, event)
	}

	return events, nil
}

// convertRecurringEvent converts a RecurringEventInput to a CalendarEvent
func (p *Parser) convertRecurringEvent(re RecurringEventInput) (models.CalendarEvent, error) {
	// Determine start date (use today if not specified)
	var startDate time.Time
	var err error
	
	if re.StartDate != "" {
		startDate, err = p.TimeParser.ParseDate(re.StartDate)
		if err != nil {
			return models.CalendarEvent{}, fmt.Errorf("failed to parse start date: %w", err)
		}
	} else {
		startDate = time.Now().In(p.TimeParser.Location)
		// Reset to start of day
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, p.TimeParser.Location)
	}

	// Parse start time
	startHour, startMin, err := p.TimeParser.ParseTime(re.StartTime)
	if err != nil {
		return models.CalendarEvent{}, fmt.Errorf("failed to parse start time: %w", err)
	}
	startTime := p.TimeParser.CombineDateTime(startDate, startHour, startMin)

	// Parse end time or calculate from duration
	var endTime time.Time
	if re.EndTime != "" {
		endHour, endMin, err := p.TimeParser.ParseTime(re.EndTime)
		if err != nil {
			return models.CalendarEvent{}, fmt.Errorf("failed to parse end time: %w", err)
		}
		endTime = p.TimeParser.CombineDateTime(startDate, endHour, endMin)
		if endTime.Before(startTime) {
			endTime = endTime.AddDate(0, 0, 1)
		}
	} else if re.Duration != "" {
		duration, err := p.TimeParser.ParseDuration(re.Duration)
		if err != nil {
			return models.CalendarEvent{}, fmt.Errorf("failed to parse duration: %w", err)
		}
		endTime = startTime.Add(duration)
	} else {
		// Default 1 hour
		endTime = startTime.Add(time.Hour)
	}

	// Convert recurrence rule
	recurrence, err := p.convertRecurrenceRule(re.Recurrence)
	if err != nil {
		return models.CalendarEvent{}, fmt.Errorf("failed to parse recurrence rule: %w", err)
	}

	return models.CalendarEvent{
		Name:        re.Name,
		Description: re.Description,
		StartTime:   startTime,
		EndTime:     endTime,
		Location:    re.Location,
		Links:       re.Links,
		Recurrence:  recurrence,
		ColorID:     re.ColorID,
	}, nil
}

// convertRecurrenceRule converts the input recurrence to a RecurrenceRule
func (p *Parser) convertRecurrenceRule(ri RecurrenceInput) (*models.RecurrenceRule, error) {
	frequency := strings.ToUpper(ri.Frequency)
	validFrequencies := map[string]bool{
		"DAILY":  true,
		"WEEKLY": true,
		"MONTHLY": true,
		"YEARLY": true,
	}

	if !validFrequencies[frequency] {
		return nil, fmt.Errorf("invalid frequency: %s", ri.Frequency)
	}

	rule := &models.RecurrenceRule{
		Frequency:       frequency,
		Interval:        ri.Interval,
		Count:           ri.Count,
		ExcludeWeekends: ri.ExcludeWeekends,
	}

	if rule.Interval == 0 {
		rule.Interval = 1
	}

	// Parse until date
	if ri.Until != "" {
		until, err := p.TimeParser.ParseDate(ri.Until)
		if err != nil {
			return nil, fmt.Errorf("failed to parse until date: %w", err)
		}
		// Set to end of day
		until = time.Date(until.Year(), until.Month(), until.Day(), 23, 59, 59, 0, p.TimeParser.Location)
		rule.Until = &until
	}

	// Normalize day names
	for i, day := range ri.ByDay {
		rule.ByDay = append(rule.ByDay, normalizeDay(day))
		_ = i
	}

	// Handle exclude_weekends by setting ByDay
	if ri.ExcludeWeekends && len(rule.ByDay) == 0 {
		rule.ByDay = []string{"MO", "TU", "WE", "TH", "FR"}
	}

	return rule, nil
}

// normalizeDay converts day names to two-letter format
func normalizeDay(day string) string {
	day = strings.ToUpper(strings.TrimSpace(day))
	
	dayMap := map[string]string{
		"MONDAY":    "MO",
		"TUESDAY":   "TU",
		"WEDNESDAY": "WE",
		"THURSDAY":  "TH",
		"FRIDAY":    "FR",
		"SATURDAY":  "SA",
		"SUNDAY":    "SU",
		"MON":       "MO",
		"TUE":       "TU",
		"WED":       "WE",
		"THU":       "TH",
		"FRI":       "FR",
		"SAT":       "SA",
		"SUN":       "SU",
	}

	if normalized, ok := dayMap[day]; ok {
		return normalized
	}
	
	// Already in short format or unknown
	return day
}
