package security

import (
	"fmt"
	"sort"
	"sync"
)

type AdmissionWindow struct {
	mu       sync.Mutex
	limits   map[string]int
	accepted map[string]int
	denied   map[string]int
}

func NewAdmissionWindow() *AdmissionWindow {
	return &AdmissionWindow{limits: make(map[string]int), accepted: make(map[string]int), denied: make(map[string]int)}
}

func (w *AdmissionWindow) SetLimit(channel string, limit int) error {
	if channel == "" || limit < 1 {
		return fmt.Errorf("admission limit is invalid")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.limits[channel] = limit
	return nil
}

func (w *AdmissionWindow) Admit(channel string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	limit, configured := w.limits[channel]
	if !configured {
		return fmt.Errorf("admission limit is not configured")
	}
	if w.accepted[channel] >= limit {
		w.denied[channel]++
		return fmt.Errorf("admission window is full")
	}
	w.accepted[channel]++
	return nil
}

func (w *AdmissionWindow) Release(channel string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.accepted[channel] < 1 {
		return false
	}
	w.accepted[channel]--
	return true
}

func (w *AdmissionWindow) Reset(channel string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.accepted, channel)
	delete(w.denied, channel)
}

func (w *AdmissionWindow) Snapshot(channel string) (int, int, int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.limits[channel], w.accepted[channel], w.denied[channel]
}

func (w *AdmissionWindow) Channels() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	channels := make([]string, 0, len(w.limits))
	for channel := range w.limits {
		channels = append(channels, channel)
	}
	sort.Strings(channels)
	return channels
}

func (w *AdmissionWindow) Available(channel string) int {
	limit, accepted, _ := w.Snapshot(channel)
	available := limit - accepted
	if available < 0 {
		return 0
	}
	return available
}
