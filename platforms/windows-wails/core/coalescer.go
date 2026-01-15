package core

import (
	"strings"
	"sync"
	"time"
)

// PendingReplace holds a pending text replacement operation
type PendingReplace struct {
	text       string
	backspaces int
	method     InjectionMethod
}

// Coalescer batches rapid diacritic changes to reduce flicker
type Coalescer struct {
	pending  *PendingReplace
	timer    *time.Timer
	mu       sync.Mutex
	apps     map[string]bool
	timerMs  int
	sendFunc func(text string, backspaces int, method InjectionMethod)
}

// NewCoalescer creates a new Coalescer with default 25ms timer
func NewCoalescer(sendFunc func(text string, backspaces int, method InjectionMethod)) *Coalescer {
	return &Coalescer{
		timerMs:  25,
		apps:     make(map[string]bool),
		sendFunc: sendFunc,
	}
}

// Queue adds a replacement to the pending queue
// If a pending replacement exists, it is replaced entirely (not accumulated)
func (c *Coalescer) Queue(text string, backspaces int, method InjectionMethod) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.timer != nil {
		c.timer.Stop()
	}

	c.pending = &PendingReplace{
		text:       text,
		backspaces: backspaces,
		method:     method,
	}

	c.timer = time.AfterFunc(time.Duration(c.timerMs)*time.Millisecond, func() {
		c.sendPending()
	})
}

// Flush immediately sends any pending replacement
func (c *Coalescer) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}

	if c.pending != nil {
		c.sendFunc(c.pending.text, c.pending.backspaces, c.pending.method)
		c.pending = nil
	}
}

// sendPending sends the pending replacement and clears it
func (c *Coalescer) sendPending() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.pending == nil {
		return
	}

	c.sendFunc(c.pending.text, c.pending.backspaces, c.pending.method)
	c.pending = nil
	c.timer = nil
}

// SetApps sets the list of apps that should use coalescing
func (c *Coalescer) SetApps(apps []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.apps = make(map[string]bool, len(apps))
	for _, app := range apps {
		c.apps[strings.ToLower(app)] = true
	}
}

// IsCoalescingApp checks if the given process name should use coalescing
func (c *Coalescer) IsCoalescingApp(currentProcessName string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.apps[strings.ToLower(currentProcessName)]
}
