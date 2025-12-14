package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/monil/calendar-event-generator/calendar"
	"github.com/monil/calendar-event-generator/config"
	"github.com/monil/calendar-event-generator/exporter"
	"github.com/monil/calendar-event-generator/interactive"
	"github.com/monil/calendar-event-generator/templates"
	"github.com/monil/calendar-event-generator/utils"
	"github.com/spf13/cobra"
)

var (
	cfg     = config.DefaultConfig()
	Version = "1.0.0"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "calendar-event-generator",
	Short: "Generate Google Calendar events from JSON templates",
	Long: `A flexible tool to create Google Calendar events from various JSON template formats.

Supported formats:
  - weekly:    Week-based schedules with events grouped by week
  - single:    Simple one-off events
  - recurring: Events with recurrence rules (daily, weekly, monthly)
  - daterange: Multi-day or all-day events

The format is auto-detected by default, or can be specified with --format.`,
	Version: Version,
	RunE: func(cmd *cobra.Command, args []string) error {
		return interactive.Run(cfg)
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add events from a JSON template to Google Calendar",
	Long:  `Parse a JSON template file and create events in Google Calendar.`,
	RunE:  runAdd,
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a JSON template without creating events",
	Long:  `Parse and validate a JSON template file without making any API calls.`,
	RunE:  runValidate,
}

var listCalendarsCmd = &cobra.Command{
	Use:   "list-calendars",
	Short: "List available Google Calendars",
	Long:  `Display all calendars available in your Google account.`,
	RunE:  runListCalendars,
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export events to an ICS file",
	Long:  `Generate an iCalendar (.ics) file from a JSON template.`,
	RunE:  runExport,
}


var inputFile string
var outputFile string
var formatOverride string

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfg.CredentialsPath, "credentials", cfg.CredentialsPath, "Path to Google OAuth credentials.json")
	rootCmd.PersistentFlags().StringVar(&cfg.TokenPath, "token", cfg.TokenPath, "Path to store OAuth token")
	rootCmd.PersistentFlags().StringVar(&cfg.CalendarID, "calendar", cfg.CalendarID, "Target calendar ID or 'primary'")
	rootCmd.PersistentFlags().StringVar(&cfg.Timezone, "timezone", cfg.Timezone, "Timezone for events (e.g., 'America/New_York', 'local')")
	rootCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", cfg.Verbose, "Enable verbose output")

	// Add command flags
	addCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input JSON template file (required)")
	addCmd.Flags().StringVarP(&formatOverride, "format", "f", "auto", "Template format: auto, weekly, single, recurring, daterange")
	addCmd.Flags().BoolVar(&cfg.DryRun, "dry-run", false, "Preview events without creating them")
	addCmd.MarkFlagRequired("input")

	// Validate command flags
	validateCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input JSON template file (required)")
	validateCmd.Flags().StringVarP(&formatOverride, "format", "f", "auto", "Template format: auto, weekly, single, recurring, daterange")
	validateCmd.MarkFlagRequired("input")

	// Register commands
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(listCalendarsCmd)
	rootCmd.AddCommand(exportCmd)

	// Export command flags
	exportCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input JSON template file (required)")
	exportCmd.Flags().StringVarP(&outputFile, "output", "o", "events.ics", "Output ICS file path")
	exportCmd.Flags().StringVarP(&formatOverride, "format", "f", "auto", "Template format: auto, weekly, single, recurring, daterange")
	exportCmd.MarkFlagRequired("input")
}

func runAdd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse template
	parser, err := templates.NewParser(cfg.Timezone)
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	format := templates.TemplateFormat(strings.ToLower(formatOverride))
	events, err := parser.ParseFile(inputFile, format)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	fmt.Printf("found %d events in template\n", len(events))

	if cfg.DryRun {
		fmt.Println("\n[DRY RUN] - No events will be created")
		utils.PrintEventSummary(events, cfg.Verbose)
		return nil
	}

	// Create calendar client
	client, err := calendar.NewClient(ctx, cfg.CredentialsPath, cfg.TokenPath, cfg.CalendarID)
	if err != nil {
		return fmt.Errorf("failed to create calendar client: %w", err)
	}

	fmt.Printf("Adding events to calendar: %s\n\n", client.GetCalendarID())

	// Create events with progress
	results, err := client.CreateEvents(events, func(current, total int, result *calendar.EventResult) {
		if result.Success {
			fmt.Printf("[OK] [%d/%d] %s\n", current, total, result.Event.Name)
			if cfg.Verbose && result.Link != "" {
				fmt.Printf("   └─ %s\n", result.Link)
			}
		} else {
			fmt.Printf("[ERR] [%d/%d] %s: %v\n", current, total, result.Event.Name, result.Error)
		}
	})

	if err != nil {
		return err
	}

	// Summary
	var successCount, failCount int
	for _, r := range results {
		if r.Success {
			successCount++
		} else {
			failCount++
		}
	}

	fmt.Printf("\nDone! Created %d events", successCount)
	if failCount > 0 {
		fmt.Printf(" (%d failed)", failCount)
	}
	fmt.Println()

	return nil
}

func runValidate(cmd *cobra.Command, args []string) error {
	parser, err := templates.NewParser(cfg.Timezone)
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	format := templates.TemplateFormat(strings.ToLower(formatOverride))
	events, err := parser.ParseFile(inputFile, format)
	if err != nil {
		return fmt.Errorf("Validation failed: %w", err)
	}

	fmt.Printf("Template is valid!\n")
	fmt.Printf("Found %d events\n\n", len(events))

	utils.PrintEventSummary(events, cfg.Verbose)

	return nil
}

func runListCalendars(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	client, err := calendar.NewClient(ctx, cfg.CredentialsPath, cfg.TokenPath, "primary")
	if err != nil {
		return fmt.Errorf("failed to create calendar client: %w", err)
	}

	calendars, err := client.ListCalendars()
	if err != nil {
		return fmt.Errorf("failed to list calendars: %w", err)
	}

	fmt.Println("Available Calendars:")
	fmt.Println("-------------------")
	for _, cal := range calendars {
		primary := ""
		if cal.Primary {
			primary = " (primary)"
		}
		fmt.Printf("  * %s%s\n", cal.Summary, primary)
		if cfg.Verbose {
			fmt.Printf("    ID: %s\n", cal.Id)
		}
	}

	return nil
}

func runExport(cmd *cobra.Command, args []string) error {
	parser, err := templates.NewParser(cfg.Timezone)
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	format := templates.TemplateFormat(strings.ToLower(formatOverride))
	events, err := parser.ParseFile(inputFile, format)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	fmt.Printf("Found %d events in template\n", len(events))

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := exporter.GenerateICS(events, f); err != nil {
		return fmt.Errorf("failed to generate ICS: %w", err)
	}

	fmt.Printf("Successfully exported to %s\n", outputFile)
	return nil
}

