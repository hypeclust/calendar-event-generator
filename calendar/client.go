package calendar

import (
	"context"
	"fmt"

	"google.golang.org/api/calendar/v3"
)

// Client wraps the Google Calendar service with helper methods
type Client struct {
	service    *calendar.Service
	calendarID string
}

// NewClient creates a new Calendar client
func NewClient(ctx context.Context, credentialsPath, tokenPath, calendarID string) (*Client, error) {
	srv, err := GetCalendarService(ctx, credentialsPath, tokenPath)
	if err != nil {
		return nil, err
	}

	if calendarID == "" {
		calendarID = "primary"
	}

	return &Client{
		service:    srv,
		calendarID: calendarID,
	}, nil
}

// ListCalendars returns all available calendars
func (c *Client) ListCalendars() ([]*calendar.CalendarListEntry, error) {
	list, err := c.service.CalendarList.List().Do()
	if err != nil {
		return nil, fmt.Errorf("unable to list calendars: %w", err)
	}
	return list.Items, nil
}

// GetCalendarID returns the current calendar ID
func (c *Client) GetCalendarID() string {
	return c.calendarID
}

// SetCalendarID sets the target calendar ID
func (c *Client) SetCalendarID(calendarID string) {
	c.calendarID = calendarID
}

// FindCalendarByName finds a calendar by its summary (name)
func (c *Client) FindCalendarByName(name string) (*calendar.CalendarListEntry, error) {
	calendars, err := c.ListCalendars()
	if err != nil {
		return nil, err
	}

	for _, cal := range calendars {
		if cal.Summary == name {
			return cal, nil
		}
	}

	return nil, fmt.Errorf("calendar not found: %s", name)
}

// GetService returns the underlying calendar service
func (c *Client) GetService() *calendar.Service {
	return c.service
}
