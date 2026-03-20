package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const timeLayout = "2006-01-02 15:04"

type Entry struct {
	ID   string     `json:"id"`
	In   time.Time  `json:"in"`
	Out  *time.Time `json:"out,omitempty"`
	Note string     `json:"note,omitempty"`
}

func (e Entry) Duration() time.Duration {
	if e.Out == nil {
		return 0
	}
	return e.Out.Sub(e.In)
}

func (e Entry) IsOpen() bool { return e.Out == nil }

type Store struct {
	path    string
	Entries []Entry `json:"entries"`
}

func Load() (*Store, error) {
	path, err := dataPath()
	if err != nil {
		return nil, err
	}
	s := &Store{path: path}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	return s, json.Unmarshal(data, s)
}

func (s *Store) Save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *Store) Add(e Entry) {
	s.Entries = append(s.Entries, e)
	sort.Slice(s.Entries, func(i, j int) bool {
		return s.Entries[i].In.Before(s.Entries[j].In)
	})
}

func (s *Store) AddOrMerge(e Entry) bool {
	ey, em, ed := e.In.Date()
	dur := e.Duration()

	for i, existing := range s.Entries {
		if existing.Note != "jira-import" {
			continue
		}
		xy, xm, xd := existing.In.Date()
		if xy == ey && xm == em && xd == ed {
			if existing.Out != nil && dur > 0 {
				newOut := existing.Out.Add(dur)
				s.Entries[i].Out = &newOut
			}
			return true
		}
	}

	s.Add(e)
	return false
}

func (s *Store) Delete(id string) bool {
	for i, e := range s.Entries {
		if e.ID == id {
			s.Entries = append(s.Entries[:i], s.Entries[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Store) FindOpenEntry() *Entry {
	for i := len(s.Entries) - 1; i >= 0; i-- {
		if s.Entries[i].IsOpen() {
			return &s.Entries[i]
		}
	}
	return nil
}

func (s *Store) CloseOpenEntry(out time.Time) (Entry, bool) {
	for i := len(s.Entries) - 1; i >= 0; i-- {
		if s.Entries[i].IsOpen() {
			s.Entries[i].Out = &out
			return s.Entries[i], true
		}
	}
	return Entry{}, false
}

func (s *Store) ByDate(date time.Time) []Entry {
	y, m, d := date.Date()
	var out []Entry
	for _, e := range s.Entries {
		ey, em, ed := e.In.Date()
		if ey == y && em == m && ed == d {
			out = append(out, e)
		}
	}
	return out
}

func (s *Store) ByWeek(date time.Time) []Entry {
	year, week := date.ISOWeek()
	var out []Entry
	for _, e := range s.Entries {
		ey, ew := e.In.ISOWeek()
		if ey == year && ew == week {
			out = append(out, e)
		}
	}
	return out
}

func (s *Store) ByMonth(date time.Time) []Entry {
	y, m, _ := date.Date()
	var out []Entry
	for _, e := range s.Entries {
		ey, em, _ := e.In.Date()
		if ey == y && em == m {
			out = append(out, e)
		}
	}
	return out
}

func ParseTime(s string) (time.Time, error) {
	if s == "now" {
		return time.Now().Truncate(time.Minute), nil
	}

	if t, err := time.ParseInLocation(timeLayout, s, time.Local); err == nil {
		return t, nil
	}

	t, err := time.ParseInLocation("15:04", s, time.Local)
	if err != nil {
		return time.Time{}, errors.New("formato inválido, use HH:MM, YYYY-MM-DD HH:MM ou \"now\"")
	}

	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, time.Local), nil
}

func dataPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".hourly")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "data.json"), nil
}

func (s *Store) DeleteAll() {
	s.Entries = []Entry{}
}

func NewID() string {
	return time.Now().Format("20060102150405")
}
