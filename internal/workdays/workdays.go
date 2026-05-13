package workdays

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type HolidayProvider interface {
	IsHoliday(t time.Time) bool
}

type brasilAPIProvider struct {
	mu    sync.Mutex
	cache map[int]map[string]struct{}
}

var DefaultHolidayProvider HolidayProvider = &brasilAPIProvider{
	cache: make(map[int]map[string]struct{}),
}

func (b *brasilAPIProvider) IsHoliday(t time.Time) bool {
	dates, err := b.fetchYear(t.Year())
	if err != nil {
		return false
	}
	_, ok := dates[t.Format("2006-01-02")]
	return ok
}

func (b *brasilAPIProvider) fetchYear(year int) (map[string]struct{}, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if dates, ok := b.cache[year]; ok {
		return dates, nil
	}

	resp, err := http.Get(fmt.Sprintf("https://brasilapi.com.br/api/feriados/v1/%d", year))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []struct {
		Date string `json:"date"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	dates := make(map[string]struct{}, len(result))
	for _, h := range result {
		dates[h.Date] = struct{}{}
	}

	b.cache[year] = dates
	return dates, nil
}

func isWorkday(t time.Time, p HolidayProvider) bool {
	w := t.Weekday()
	if w == time.Saturday || w == time.Sunday {
		return false
	}
	return !p.IsHoliday(t)
}

func countDays(from, to time.Time, p HolidayProvider) int {
	count := 0
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		if isWorkday(d, p) {
			count++
		}
	}
	return count
}

func monthStart(ref time.Time) time.Time {
	return time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, ref.Location())
}

func monthEnd(ref time.Time) time.Time {
	return time.Date(ref.Year(), ref.Month()+1, 0, 0, 0, 0, 0, ref.Location())
}

func IsWorkday(t time.Time, p ...HolidayProvider) bool {
	return isWorkday(t, resolveProvider(p))
}

func CountInMonth(ref time.Time, p ...HolidayProvider) int {
	return countDays(monthStart(ref), monthEnd(ref), resolveProvider(p))
}

func CountUntilToday(ref time.Time, p ...HolidayProvider) int {
	today := time.Date(ref.Year(), ref.Month(), ref.Day(), 0, 0, 0, 0, ref.Location())
	return countDays(monthStart(ref), today, resolveProvider(p))
}

func CountFromTomorrow(ref time.Time, p ...HolidayProvider) int {
	tomorrow := time.Date(ref.Year(), ref.Month(), ref.Day()+1, 0, 0, 0, 0, ref.Location())
	return countDays(tomorrow, monthEnd(ref), resolveProvider(p))
}

func resolveProvider(providers []HolidayProvider) HolidayProvider {
	if len(providers) > 0 && providers[0] != nil {
		return providers[0]
	}
	return DefaultHolidayProvider
}
