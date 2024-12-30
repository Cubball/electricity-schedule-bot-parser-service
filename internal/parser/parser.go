package parser

import (
	"context"
	"electricity-schedule-bot/parser-service/internal/models"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

const (
	digits     = "0123456789"
	dateLayout = "02.01.2006"
	timeLayout = "15:04"
)

var QueueNumbers = []string{
	"1.1",
	"1.2",
	"2.1",
	"2.2",
	"3.1",
	"3.2",
	"4.1",
	"4.2",
	"5.1",
	"5.2",
	"6.1",
	"6.2",
}

func Parse(ctx context.Context, webPage *goquery.Document) (*models.Schedule, error) {
	selection := webPage.Find("tbody > tr")
	trElems := slices.Clone(selection.Nodes)
	if len(trElems) == 0 {
		return nil, errors.New("no `tr` elements were found on the web page")
	}

    slog.DebugContext(ctx, "found tr elements on the web page", "trElemsCount", len(trElems))
	scheduleEntries := []models.ScheduleEntry{}
	slices.Reverse(trElems)
	for _, trElem := range trElems {
		tdElems := slices.Collect(trElem.ChildNodes())
		if !trHasScheduleEntries(tdElems) {
            slog.DebugContext(ctx, "found the first tr to not contain schedule entries, breaking")
			break
		}

		trEntries, err := parseTr(ctx, tdElems)
		if err != nil {
			return nil, err
		}

		scheduleEntries = append(scheduleEntries, trEntries...)
	}

	return &models.Schedule{
		Entries:   scheduleEntries,
		FetchTime: time.Now().UTC(), // TODO: inject time provider?
	}, nil
}

func trHasScheduleEntries(tdElems []*html.Node) bool {
	if len(tdElems) == 0 {
		return false
	}

	firstChildContent := goquery.NewDocumentFromNode(tdElems[0])
	return strings.ContainsAny(firstChildContent.Text(), digits)
}

func parseTr(ctx context.Context, tdElems []*html.Node) ([]models.ScheduleEntry, error) {
    slog.DebugContext(ctx, "parsing tr", "tdElemsCount", len(tdElems))
	firstChildContent := goquery.NewDocumentFromNode(tdElems[0])
	date, err := time.Parse(dateLayout, firstChildContent.Text())
	if err != nil {
		return nil, fmt.Errorf("failed to parse the date for `tr`: %w", err)
	}

    slog.DebugContext(ctx, "parsed the date in a tr", "date", date)
	scheduleEntries := []models.ScheduleEntry{}
	for idx, tdElem := range tdElems[1:] {
		if idx >= len(QueueNumbers) {
			return nil, fmt.Errorf("too many `td` elements (except the first one): expected %d, got %d", len(QueueNumbers), len(tdElems)-1)
		}

		text := getTextFromTd(tdElem)
        slog.DebugContext(ctx, "parsing the content of a td", "content", text)
		tdEntries, err := parseTdText(ctx, date, QueueNumbers[idx], text)
		if err != nil {
			return nil, err
		}

        slog.DebugContext(ctx, "parsed schedule entries from a td", "scheduleEntriesCount", len(tdEntries))
		scheduleEntries = append(scheduleEntries, tdEntries...)
	}

	return scheduleEntries, nil
}

func getTextFromTd(tdElem *html.Node) string {
	contents := []string{}
	for child := range tdElem.ChildNodes() {
		document := goquery.NewDocumentFromNode(child)
		contents = append(contents, document.Text())
	}

	return strings.Join(contents, "\n")
}

func parseTdText(ctx context.Context, date time.Time, queueNumber, text string) ([]models.ScheduleEntry, error) {
	timePeriods := strings.Split(text, "\n")
    slog.DebugContext(ctx, "split up time periods in the td text", "timePeriodsCount", len(timePeriods))
	scheduleEntries := []models.ScheduleEntry{}
	for _, timePeriod := range timePeriods {
		if !strings.ContainsAny(timePeriod, digits) {
            slog.DebugContext(ctx, "time period does not contain digits, skipping", "content", timePeriod)
			continue
		}

		parts := strings.Split(timePeriod, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("unexpected number of parts in the time period %q, expected 2, got %d", timePeriod, len(parts))
		}

		start, err := time.Parse(timeLayout, strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("failed to parse the start time of the time period: %q. %w", parts[0], err)
		}

		end, err := time.Parse(timeLayout, strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("failed to parse the end time of the time period: %q. %w", parts[1], err)
		}

        slog.DebugContext(ctx, "parsed the time period", "startTime", start, "endTime", end)
		scheduleEntries = append(scheduleEntries, models.ScheduleEntry{
			QueueNumber: queueNumber,
			Date:        models.DateOnly(date),
			StartTime:   models.TimeOnly(start),
			EndTime:     models.TimeOnly(end),
		})
	}

	return scheduleEntries, nil
}
