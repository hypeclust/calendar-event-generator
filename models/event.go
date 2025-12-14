package models

import "time"

// CalendarEvent represents a unified calendar event structure
// that can be created from any supported template format
type CalendarEvent struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     time.Time      `json:"end_time"`
	Location    string         `json:"location,omitempty"`
	Links       []string       `json:"links,omitempty"`
	AllDay      bool           `json:"all_day,omitempty"`
	Recurrence  *RecurrenceRule `json:"recurrence,omitempty"`
	Reminders   []Reminder     `json:"reminders,omitempty"`
	ColorID     string         `json:"color_id,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// RecurrenceRule defines how an event should repeat
type RecurrenceRule struct {
	Frequency       string     `json:"frequency"` // DAILY, WEEKLY, MONTHLY, YEARLY
	Interval        int        `json:"interval"`  // Every N frequency units
	Until           *time.Time `json:"until,omitempty"`
	Count           int        `json:"count,omitempty"`    // Number of occurrences
	ByDay           []string   `json:"by_day,omitempty"`   // MO, TU, WE, TH, FR, SA, SU
	ExcludeWeekends bool       `json:"exclude_weekends,omitempty"`
}

// Reminder defines when to remind the user about an event
type Reminder struct {
	Method  string `json:"method"` // "email" or "popup"
	Minutes int    `json:"minutes"` // Minutes before event
}

// ToRRuleString converts the recurrence rule to iCalendar RRULE format
func (r *RecurrenceRule) ToRRuleString() string {
	if r == nil {
		return ""
	}

	rule := "RRULE:FREQ=" + r.Frequency

	if r.Interval > 1 {
		rule += ";INTERVAL=" + string(rune(r.Interval+'0'))
	}

	if r.Until != nil {
		rule += ";UNTIL=" + r.Until.Format("20060102T150405Z")
	}

	if r.Count > 0 {
		rule += ";COUNT=" + string(rune(r.Count+'0'))
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

// FormatDescription creates a formatted event description with links
func (e *CalendarEvent) FormatDescription() string {
	desc := e.Description

	if len(e.Links) > 0 {
		if desc != "" {
			desc += "\n\n"
		}
		desc += "Useful Links:\n"
		for _, link := range e.Links {
			desc += "- " + link + "\n"
		}
	}

	return desc
}
