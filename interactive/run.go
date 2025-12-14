package interactive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/monil/calendar-event-generator/calendar"
	"github.com/monil/calendar-event-generator/config"
	"github.com/monil/calendar-event-generator/exporter"
	"github.com/monil/calendar-event-generator/templates"
	"github.com/monil/calendar-event-generator/utils"
)

// Run starts the interactive CLI mode
func Run(cfg *config.Config) error {
	var action string
	var selectedFile string
	var calendarID string
	var dryRun bool

	// Styles (No emojis!)
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginBottom(1)

	fmt.Println(titleStyle.Render("Calendar Event Generator"))

	// 1. Choose Action
	err := huh.NewSelect[string]().
		Title("What would you like to do?").
		Options(
			huh.NewOption("Add Events from Template", "add"),
			huh.NewOption("Validate Template", "validate"),
			huh.NewOption("Export to ICS", "export"),
			huh.NewOption("List Calendars", "list"),
		).
		Value(&action).
		WithTheme(huh.ThemeBase()).
		Run()

	if err != nil {
		return err
	}

	if action == "list" {
		return runListCalendars(cfg)
	}

	// 2. Choose File
	files, err := getExampleFiles()
	if err != nil {
		return fmt.Errorf("failed to list examples: %w", err)
	}

	fileOptions := make([]huh.Option[string], len(files))
	for i, f := range files {
		fileOptions[i] = huh.NewOption(filepath.Base(f), f)
	}
	// Add an option to enter manually
	fileOptions = append(fileOptions, huh.NewOption("Enter file path manualy...", "manual"))

	err = huh.NewSelect[string]().
		Title("Select a template file").
		Options(fileOptions...).
		Value(&selectedFile).
		WithTheme(huh.ThemeBase()).
		Run()

	if err != nil {
		return err
	}

	if selectedFile == "manual" {
		err = huh.NewInput().
			Title("Enter file path").
			Value(&selectedFile).
			WithTheme(huh.ThemeBase()).
			Run()
		if err != nil {
			return err
		}
	}

	// 3. Configuration (for Add/Validate)
	if action == "add" {
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Calendar ID").
					Description("Leave empty for 'primary'").
					Value(&calendarID),
				huh.NewConfirm().
					Title("Dry Run?").
					Description("Preview without creating events").
					Value(&dryRun),
			),
		).WithTheme(huh.ThemeBase())

		err = form.Run()
		if err != nil {
			return err
		}

		if calendarID != "" {
			cfg.CalendarID = calendarID
		}
		cfg.DryRun = dryRun

		return runAdd(cfg, selectedFile)
	} else if action == "validate" {
		return runValidate(cfg, selectedFile)
	} else if action == "export" {
		return runExport(cfg, selectedFile)
	}

	return nil
}

func getExampleFiles() ([]string, error) {
	var files []string
	entries, err := os.ReadDir("examples")
	if err != nil {
		// If examples dir doesn't exist, just return empty, don't fail
		return []string{}, nil
	}

	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			files = append(files, filepath.Join("examples", e.Name()))
		}
	}
	return files, nil
}

// Logic duplicated/adapted from main.go to avoid import cycle or complex refactor
// Ideally this logic should be in a 'usecase' package.

func runListCalendars(cfg *config.Config) error {
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
	for _, cal := range calendars {
		primary := ""
		if cal.Primary {
			primary = " (primary)"
		}
		fmt.Printf("  * %s%s\n", cal.Summary, primary)
	}
	return nil
}

func runAdd(cfg *config.Config, inputFile string) error {
	// Parse template
	parser, err := templates.NewParser(cfg.Timezone)
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	// Auto-detect format
	events, err := parser.ParseFile(inputFile, templates.FormatAuto)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	fmt.Printf("\nFound %d events in template\n", len(events))

	if cfg.DryRun {
		fmt.Println("\n[DRY RUN] - No events will be created")
		utils.PrintEventSummary(events, cfg.Verbose)

		var confirm bool
		err := huh.NewConfirm().
			Title("Do you want to proceed with adding these events?").
			Value(&confirm).
			WithTheme(huh.ThemeBase()).
			Run()
		if err != nil {
			return err
		}

		if !confirm {
			return nil
		}
		// Proceed with adding events
		fmt.Println()
	}

	// Create calendar client
	ctx := context.Background()
	client, err := calendar.NewClient(ctx, cfg.CredentialsPath, cfg.TokenPath, cfg.CalendarID)
	if err != nil {
		return fmt.Errorf("failed to create calendar client: %w", err)
	}

	fmt.Printf("Adding events to calendar: %s\n\n", client.GetCalendarID())

	// Create events with spinner/progress
	// Huh doesn't have a progress bar yet, but we can just print simple logs
	_, err = client.CreateEvents(events, func(current, total int, result *calendar.EventResult) {
		if result.Success {
			fmt.Printf("[OK] [%d/%d] %s\n", current, total, result.Event.Name)
		} else {
			fmt.Printf("[ERR] [%d/%d] %s: %v\n", current, total, result.Event.Name, result.Error)
		}
	})

	if err != nil {
		return err
	}

	fmt.Println("\nDone!")
	return nil
}

func runValidate(cfg *config.Config, inputFile string) error {
	parser, err := templates.NewParser(cfg.Timezone)
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	events, err := parser.ParseFile(inputFile, templates.FormatAuto)
	if err != nil {
		return fmt.Errorf("Validation failed: %w", err)
	}

	fmt.Println("Template is valid!")
	fmt.Printf("Found %d events\n", len(events))
	return nil
}

func runExport(cfg *config.Config, inputFile string) error {
	var outputFile string

	err := huh.NewInput().
		Title("Output File Path").
		Value(&outputFile).
		WithTheme(huh.ThemeBase()).
		Run()
	if err != nil {
		return err
	}

	if outputFile == "" {
		outputFile = "events.ics"
	}

	// Ensure .ics extension
	if !strings.HasSuffix(outputFile, ".ics") {
		outputFile += ".ics"
	}

	parser, err := templates.NewParser(cfg.Timezone)
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	events, err := parser.ParseFile(inputFile, templates.FormatAuto)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := exporter.GenerateICS(events, f); err != nil {
		return fmt.Errorf("failed to generate ICS: %w", err)
	}

	fmt.Printf("\nSuccessfully exported %d events to %s\n", len(events), outputFile)
	return nil
}
