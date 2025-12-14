package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TimeParser handles parsing of various time and date formats
type TimeParser struct {
	Location *time.Location
}

// NewTimeParser creates a new TimeParser with the given timezone
func NewTimeParser(timezone string) (*TimeParser, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}
	return &TimeParser{Location: loc}, nil
}

// NewTimeParserLocal creates a TimeParser using the local timezone
func NewTimeParserLocal() *TimeParser {
	return &TimeParser{Location: time.Local}
}

// ParseDate parses various date formats and returns a time.Time
// Supported formats:
// - 2025-12-09 (ISO)
// - 12/09/2025 (US)
// - 09-12-2025 (EU)
// - December 9, 2025
func (tp *TimeParser) ParseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	
	formats := []string{
		"2006-01-02",          // ISO
		"01/02/2006",          // US
		"02-01-2006",          // EU
		"January 2, 2006",     // Long format
		"Jan 2, 2006",         // Short month
		"2006/01/02",          // Alternative ISO
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, dateStr, tp.Location); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// ParseTime parses various time formats
// Supported formats:
// - 10:00 (24h)
// - 10:00:00 (24h with seconds)
// - 10:00am, 10:00 AM (12h)
// - 10:00 a.m.
func (tp *TimeParser) ParseTime(timeStr string) (hour, minute int, err error) {
	timeStr = strings.TrimSpace(timeStr)
	timeStr = strings.ToLower(timeStr)
	timeStr = strings.ReplaceAll(timeStr, ".", "")
	timeStr = strings.ReplaceAll(timeStr, " ", "")

	// Check for AM/PM
	isPM := strings.Contains(timeStr, "pm")
	isAM := strings.Contains(timeStr, "am")
	timeStr = strings.ReplaceAll(timeStr, "pm", "")
	timeStr = strings.ReplaceAll(timeStr, "am", "")

	// Parse hours and minutes
	parts := strings.Split(timeStr, ":")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hour, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid hour: %s", parts[0])
	}

	minute, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minute: %s", parts[1])
	}

	// Convert 12h to 24h format
	if isPM && hour < 12 {
		hour += 12
	} else if isAM && hour == 12 {
		hour = 0
	}

	return hour, minute, nil
}

// ParseTimeRange parses time ranges like "11:00am – 1:00pm" or "10:00 - 12:00"
// Returns start and end times as hours and minutes
func (tp *TimeParser) ParseTimeRange(rangeStr string) (startHour, startMin, endHour, endMin int, err error) {
	// Normalize separators
	rangeStr = strings.ReplaceAll(rangeStr, "–", "-") // em dash to hyphen
	rangeStr = strings.ReplaceAll(rangeStr, "—", "-") // en dash to hyphen
	rangeStr = strings.ReplaceAll(rangeStr, "to", "-")
	rangeStr = strings.ReplaceAll(rangeStr, "TO", "-")

	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return 0, 0, 0, 0, fmt.Errorf("invalid time range format: %s", rangeStr)
	}

	startHour, startMin, err = tp.ParseTime(parts[0])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid start time: %w", err)
	}

	endHour, endMin, err = tp.ParseTime(parts[1])
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("invalid end time: %w", err)
	}

	return startHour, startMin, endHour, endMin, nil
}

// ParseDuration parses duration strings like "2h", "30m", "1h30m", "90min"
func (tp *TimeParser) ParseDuration(durationStr string) (time.Duration, error) {
	durationStr = strings.TrimSpace(strings.ToLower(durationStr))

	// Handle standard Go duration format
	if d, err := time.ParseDuration(durationStr); err == nil {
		return d, nil
	}

	// Handle alternative formats
	durationStr = strings.ReplaceAll(durationStr, "hr", "h")
	durationStr = strings.ReplaceAll(durationStr, "hour", "h")
	durationStr = strings.ReplaceAll(durationStr, "hours", "h")
	durationStr = strings.ReplaceAll(durationStr, "min", "m")
	durationStr = strings.ReplaceAll(durationStr, "minute", "m")
	durationStr = strings.ReplaceAll(durationStr, "minutes", "m")

	if d, err := time.ParseDuration(durationStr); err == nil {
		return d, nil
	}

	// Try to extract hours and minutes with regex
	re := regexp.MustCompile(`(\d+)\s*h(?:ours?)?\s*(?:(\d+)\s*m)?|(\d+)\s*m(?:in(?:utes?)?)?`)
	matches := re.FindStringSubmatch(durationStr)
	
	if len(matches) > 0 {
		var hours, mins int
		if matches[1] != "" {
			hours, _ = strconv.Atoi(matches[1])
		}
		if matches[2] != "" {
			mins, _ = strconv.Atoi(matches[2])
		}
		if matches[3] != "" {
			mins, _ = strconv.Atoi(matches[3])
		}
		return time.Duration(hours)*time.Hour + time.Duration(mins)*time.Minute, nil
	}

	return 0, fmt.Errorf("unable to parse duration: %s", durationStr)
}

// CombineDateTime combines a date and time components into a single time.Time
func (tp *TimeParser) CombineDateTime(date time.Time, hour, minute int) time.Time {
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		hour, minute, 0, 0,
		tp.Location,
	)
}

// ParseDateTime parses a date and time string together
func (tp *TimeParser) ParseDateTime(dateStr, timeStr string) (time.Time, error) {
	date, err := tp.ParseDate(dateStr)
	if err != nil {
		return time.Time{}, err
	}

	hour, minute, err := tp.ParseTime(timeStr)
	if err != nil {
		return time.Time{}, err
	}

	return tp.CombineDateTime(date, hour, minute), nil
}

// ParseDateTimeRange parses a date with a time range, returning start and end times
func (tp *TimeParser) ParseDateTimeRange(dateStr, timeRangeStr string) (start, end time.Time, err error) {
	date, err := tp.ParseDate(dateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	startHour, startMin, endHour, endMin, err := tp.ParseTimeRange(timeRangeStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	start = tp.CombineDateTime(date, startHour, startMin)
	end = tp.CombineDateTime(date, endHour, endMin)

	// Handle overnight events (end time is before start time)
	if end.Before(start) {
		end = end.AddDate(0, 0, 1)
	}

	return start, end, nil
}
