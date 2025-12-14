package utils

import (
	"fmt"
	"strings"

	"github.com/monil/calendar-event-generator/models"
)

// PrintEventSummary prints a formatted summary of the events to stdout
func PrintEventSummary(events []models.CalendarEvent, verbose bool) {
	fmt.Println("Events to be created:")
	fmt.Println("-------------------")

	for i, e := range events {
		timeStr := e.StartTime.Format("Mon, Jan 2 2006 3:04 PM")
		if e.AllDay {
			timeStr = e.StartTime.Format("Mon, Jan 2 2006") + " (All day)"
		}

		fmt.Printf("%3d. %s\n", i+1, e.Name)
		fmt.Printf("     Date: %s\n", timeStr)

		if e.Location != "" {
			fmt.Printf("     Loc: %s\n", e.Location)
		}

		if e.Recurrence != nil {
			fmt.Printf("     Repeats: %s", strings.ToLower(e.Recurrence.Frequency))
			if e.Recurrence.Until != nil {
				fmt.Printf(" until %s", e.Recurrence.Until.Format("Jan 2, 2006"))
			}
			fmt.Println()
		}

		if verbose && e.Description != "" {
			desc := e.Description
			if len(desc) > 100 {
				desc = desc[:100] + "..."
			}
			fmt.Printf("     Desc: %s\n", desc)
		}

		fmt.Println()
	}
}
