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

// HACK: ignoring the possible error, it shouldn't
var location, _ = time.LoadLocation("Europe/Kyiv")

var queueNumbers = []string{
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
	slices.Reverse(trElems)
	queues := make([]models.Queue, 0, len(queueNumbers))
	for _, queueNumber := range queueNumbers {
		queues = append(queues, models.Queue{
			Number: queueNumber,
		})
	}

	for _, trElem := range trElems {
		tdElems := slices.Collect(trElem.ChildNodes())
		if !trHasScheduleEntries(tdElems) {
			slog.DebugContext(ctx, "found the first tr to not contain schedule entries, breaking")
			break
		}

		err := parseTr(ctx, tdElems, queues)
		if err != nil {
			return nil, err
		}
	}

	return &models.Schedule{
		FetchTime: time.Now().UTC(), // TODO: inject time provider?
		Queues:    queues,
	}, nil
}

func trHasScheduleEntries(tdElems []*html.Node) bool {
	if len(tdElems) == 0 {
		return false
	}

	firstChildContent := goquery.NewDocumentFromNode(tdElems[0])
	return strings.ContainsAny(firstChildContent.Text(), digits)
}

func parseTr(ctx context.Context, tdElems []*html.Node, queues []models.Queue) error {
	slog.DebugContext(ctx, "parsing tr", "tdElemsCount", len(tdElems))
	firstChildContent := goquery.NewDocumentFromNode(tdElems[0])
	date, err := time.Parse(dateLayout, firstChildContent.Text())
	if err != nil {
		return fmt.Errorf("failed to parse the date for `tr`: %w", err)
	}

	slog.DebugContext(ctx, "parsed the date in a tr", "date", date)
	for idx, tdElem := range tdElems[1:] {
		if idx >= len(queues) {
			return fmt.Errorf("too many `td` elements (except the first one): expected %d, got %d", len(queues), len(tdElems)-1)
		}

        queue := queues[idx]
		text := getTextFromTd(tdElem)
		slog.DebugContext(ctx, "parsing the content of a td", "content", text)
		tdEntries, err := parseTdText(ctx, date, text)
		if err != nil {
			return err
		}

		slog.DebugContext(ctx, "parsed schedule entries from a td", "scheduleEntriesCount", len(tdEntries))
        queue.DisconnectionTimes = append(queue.DisconnectionTimes, tdEntries...)
	}

	return nil
}

func getTextFromTd(tdElem *html.Node) string {
	contents := []string{}
	for child := range tdElem.ChildNodes() {
		document := goquery.NewDocumentFromNode(child)
		contents = append(contents, document.Text())
	}

	return strings.Join(contents, "\n")
}

func parseTdText(ctx context.Context, date time.Time, text string) ([]models.DisconnectionTime, error) {
	timePeriods := strings.Split(text, "\n")
	slog.DebugContext(ctx, "split up time periods in the td text", "timePeriodsCount", len(timePeriods))
	disconnectionTimes := []models.DisconnectionTime{}
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

		start = time.Date(date.Year(), date.Month(), date.Day(), start.Hour(), start.Minute(), 0, 0, location).UTC()
		end = time.Date(date.Year(), date.Month(), date.Day(), end.Hour(), end.Minute(), 0, 0, location).UTC()
		slog.DebugContext(ctx, "parsed the time period", "startTime", start, "endTime", end)
        disconnectionTimes = append(disconnectionTimes, models.DisconnectionTime{
            Start: start,
            End: end,
        })
	}

	return disconnectionTimes, nil
}
