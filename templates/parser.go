package templates

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/monil/calendar-event-generator/models"
	"github.com/monil/calendar-event-generator/utils"
)

// TemplateFormat represents the type of JSON template
type TemplateFormat string

const (
	FormatWeekly    TemplateFormat = "weekly"
	FormatSingle    TemplateFormat = "single"
	FormatRecurring TemplateFormat = "recurring"
	FormatDateRange TemplateFormat = "daterange"
	FormatAuto      TemplateFormat = "auto"
)

// Parser is the main template parser that routes to specific parsers
type Parser struct {
	TimeParser *utils.TimeParser
}

// NewParser creates a new template parser
func NewParser(timezone string) (*Parser, error) {
	var tp *utils.TimeParser
	var err error

	if timezone == "" || timezone == "local" {
		tp = utils.NewTimeParserLocal()
	} else {
		tp, err = utils.NewTimeParser(timezone)
		if err != nil {
			return nil, err
		}
	}

	return &Parser{TimeParser: tp}, nil
}

// ParseFile reads and parses a JSON file, auto-detecting the format
func (p *Parser) ParseFile(filename string, format TemplateFormat) ([]models.CalendarEvent, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	return p.Parse(data, format)
}

// Parse parses JSON data, auto-detecting the format if not specified
func (p *Parser) Parse(data []byte, format TemplateFormat) ([]models.CalendarEvent, error) {
	if format == FormatAuto || format == "" {
		format = p.detectFormat(data)
	}

	switch format {
	case FormatWeekly:
		return p.parseWeekly(data)
	case FormatSingle:
		return p.parseSingle(data)
	case FormatRecurring:
		return p.parseRecurring(data)
	case FormatDateRange:
		return p.parseDateRange(data)
	default:
		return nil, fmt.Errorf("unknown template format: %s", format)
	}
}

// detectFormat attempts to auto-detect the JSON template format
func (p *Parser) detectFormat(data []byte) TemplateFormat {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return FormatSingle // Default fallback
	}

	// Check for explicit format field
	if formatRaw, ok := raw["format"]; ok {
		var format string
		if err := json.Unmarshal(formatRaw, &format); err == nil {
			return TemplateFormat(strings.ToLower(format))
		}
	}

	// Check for week_* keys (weekly format)
	for key := range raw {
		if strings.HasPrefix(strings.ToLower(key), "week_") ||
			strings.HasPrefix(strings.ToLower(key), "week ") ||
			strings.ToLower(key) == "week1" {
			return FormatWeekly
		}
	}

	// Check for events array with recurrence
	if eventsRaw, ok := raw["events"]; ok {
		var events []map[string]json.RawMessage
		if err := json.Unmarshal(eventsRaw, &events); err == nil && len(events) > 0 {
			// Check first event for recurrence field
			if _, hasRecurrence := events[0]["recurrence"]; hasRecurrence {
				return FormatRecurring
			}
			// Check for date range fields
			if _, hasEndDate := events[0]["end_date"]; hasEndDate {
				return FormatDateRange
			}
		}
	}

	return FormatSingle
}
