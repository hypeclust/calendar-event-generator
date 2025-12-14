# Calendar Event Generator

A cross-platform Go CLI tool to create Google Calendar events from flexible JSON templates.

## Features

- **Multiple Template Formats**: Weekly schedules, single events, recurring events, and date ranges
- **Interactive CLI**: Easy-to-use menu system for all operations
- **ICS Export**: Convert JSON templates to standard .ics files
- **Auto-Detection**: Automatically detects JSON template format
- **Dry Run Mode**: Preview events before creating them
- **Cross-Platform**: Works on Windows, macOS, and Linux
- **OAuth 2.0**: Secure authentication with Google Calendar API

## Installation

```bash
# Clone/download the project, then:
cd calendar-event-generator
go build -o calendar-event-generator.exe  # Windows
go build -o calendar-event-generator      # macOS/Linux
```

## Google Cloud Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project
3. Enable the **Google Calendar API**
4. Create OAuth 2.0 credentials (Desktop application)
5. Download and save as `credentials.json` in the project directory

## Usage

### Interactive Mode (Default)
Simply run the executable without arguments to start the interactive CLI:
```bash
./calendar-event-generator
```
This will guide you through adding events, validating templates, or exporting to ICS.

### Add Events from Template (CLI)
```bash
# Add events to your primary calendar
./calendar-event-generator add --input schedule.json

# Preview without creating (dry run)
./calendar-event-generator add --input schedule.json --dry-run
```

### Export to ICS
```bash
./calendar-event-generator export --input schedule.json --output events.ics
```

### Validate Template
```bash
./calendar-event-generator validate --input schedule.json
```

### List Calendars
```bash
./calendar-event-generator list-calendars
```

## Template Formats

### Weekly Schedule
For schedules organized by week:
```json
{
  "week_1": [
    {
      "event_name": "Study Session",
      "date": "2025-12-09",
      "time": "11:00am – 1:00pm",
      "topic_details": "Topic description",
      "useful_links": ["https://example.com"]
    }
  ]
}
```

### Single Events
For one-off events:
```json
{
  "format": "single",
  "events": [
    {
      "name": "Meeting",
      "date": "2025-12-15",
      "start_time": "10:00",
      "end_time": "11:00",
      "description": "Weekly sync",
      "location": "Room A"
    }
  ]
}
```

### Recurring Events
For repeated events:
```json
{
  "format": "recurring",
  "events": [
    {
      "name": "Daily Standup",
      "start_time": "09:00",
      "duration": "15m",
      "recurrence": {
        "frequency": "DAILY",
        "until": "2025-12-31",
        "exclude_weekends": true
      }
    }
  ]
}
```

### Date Range Events
For multi-day events:
```json
{
  "format": "daterange",
  "events": [
    {
      "name": "Conference",
      "start_date": "2025-12-20",
      "end_date": "2025-12-22",
      "all_day": true
    }
  ]
}
```

## CLI Options

```
Global Flags:
  --credentials   Path to Google OAuth credentials.json
  --token         Path to store OAuth token
  --calendar      Target calendar ID or 'primary'
  --timezone      Timezone (e.g., 'America/New_York', 'local')
  -v, --verbose   Enable verbose output

Add Command Flags:
  -i, --input     Input JSON template file (required)
  -f, --format    Template format: auto, weekly, single, recurring, daterange
  --dry-run       Preview events without creating them
```

## Cross-Platform Builds

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o calendar-event-generator.exe

# macOS
GOOS=darwin GOARCH=amd64 go build -o calendar-event-generator

# Linux
GOOS=linux GOARCH=amd64 go build -o calendar-event-generator
```

## License

MIT
