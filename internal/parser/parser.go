package parser

import (
	"electricity-schedule-bot/parser-service/internal/models"
	"errors"
	"fmt"
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

func Parse(webPage *goquery.Document) (*models.Schedule, error) {
	selection := webPage.Find("tbody > tr")
	trElems := slices.Clone(selection.Nodes)
	if len(trElems) == 0 {
		return nil, errors.New("no `tr` elements were found on the web page")
	}

	scheduleEntries := []models.ScheduleEntry{}
	slices.Reverse(trElems)
	for _, trElem := range trElems {
		tdElems := slices.Collect(trElem.ChildNodes())
		if !trHasScheduleEntries(tdElems) {
			break
		}

		trEntries, err := parseTr(tdElems)
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

func parseTr(tdElems []*html.Node) ([]models.ScheduleEntry, error) {
	firstChildContent := goquery.NewDocumentFromNode(tdElems[0])
	date, err := time.Parse(dateLayout, firstChildContent.Text())
	if err != nil {
		return nil, fmt.Errorf("failed to parse the date for `tr`: %w", err)
	}

	scheduleEntries := []models.ScheduleEntry{}
	for idx, tdElem := range tdElems[1:] {
		if idx >= len(QueueNumbers) {
			return nil, fmt.Errorf("too many `td` elements (except the first one): expected %d, got %d", len(QueueNumbers), len(tdElems)-1)
		}

		text := getTextFromTd(tdElem)
		tdEntries, err := parseTdText(date, QueueNumbers[idx], text)
		if err != nil {
			return nil, err
		}

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

func parseTdText(date time.Time, queueNumber, text string) ([]models.ScheduleEntry, error) {
	timePeriods := strings.Split(text, "\n")
	scheduleEntries := []models.ScheduleEntry{}
	for _, timePeriod := range timePeriods {
		if !strings.ContainsAny(timePeriod, digits) {
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

		scheduleEntries = append(scheduleEntries, models.ScheduleEntry{
			QueueNumber: queueNumber,
			Date:        models.DateOnly(date),
			StartTime:   models.TimeOnly(start),
			EndTime:     models.TimeOnly(end),
		})
	}

	return scheduleEntries, nil
}
