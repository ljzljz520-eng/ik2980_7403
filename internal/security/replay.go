package security

import (
	"fmt"
	"sort"
	"sync"
)

type ReplayGuard struct {
	mu   sync.Mutex
	seen map[string]map[string]struct{}
}

func NewReplayGuard() *ReplayGuard {
	return &ReplayGuard{seen: make(map[string]map[string]struct{})}
}

func (g *ReplayGuard) Reserve(channel, nonce string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if channel == "" || nonce == "" {
		return fmt.Errorf("replay key is incomplete")
	}
	if _, ok := g.seen[channel]; !ok {
		g.seen[channel] = make(map[string]struct{})
	}
	if _, exists := g.seen[channel][nonce]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateNonce, nonce)
	}
	g.seen[channel][nonce] = struct{}{}
	return nil
}

func (g *ReplayGuard) Forget(channel, nonce string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	nonces, ok := g.seen[channel]
	if !ok {
		return false
	}
	if _, ok = nonces[nonce]; !ok {
		return false
	}
	delete(nonces, nonce)
	return true
}

func (g *ReplayGuard) Contains(channel, nonce string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	_, ok := g.seen[channel][nonce]
	return ok
}

func (g *ReplayGuard) Snapshot(channel string) []string {
	g.mu.Lock()
	defer g.mu.Unlock()
	values := make([]string, 0, len(g.seen[channel]))
	for nonce := range g.seen[channel] {
		values = append(values, nonce)
	}
	sort.Strings(values)
	return values
}

func (g *ReplayGuard) Count(channel string) int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.seen[channel])
}
