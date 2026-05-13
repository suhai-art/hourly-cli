package workdays

import "time"

func IsWorkday(t time.Time) bool {
	w := t.Weekday()
	return w != time.Saturday && w != time.Sunday
}

func CountInMonth(ref time.Time) int {
	first := time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, ref.Location())
	last := first.AddDate(0, 1, -1)

	count := 0
	for d := first; !d.After(last); d = d.AddDate(0, 0, 1) {
		if IsWorkday(d) {
			count++
		}
	}
	return count
}

func CountUntilToday(ref time.Time) int {
	first := time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, ref.Location())
	today := time.Date(ref.Year(), ref.Month(), ref.Day(), 0, 0, 0, 0, ref.Location())

	count := 0
	for d := first; !d.After(today); d = d.AddDate(0, 0, 1) {
		if IsWorkday(d) {
			count++
		}
	}
	return count
}

func CountFromTomorrow(ref time.Time) int {
	tomorrow := ref.AddDate(0, 0, 1)
	last := time.Date(ref.Year(), ref.Month()+1, 0, 0, 0, 0, 0, ref.Location())

	count := 0
	for d := tomorrow; !d.After(last); d = d.AddDate(0, 0, 1) {
		if IsWorkday(d) {
			count++
		}
	}
	return count
}
