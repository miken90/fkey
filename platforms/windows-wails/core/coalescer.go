package core

import (
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
	timerMs  int
	sendFunc func(text string, backspaces int, method InjectionMethod)
}

// NewCoalescer creates a new Coalescer with default 25ms timer
func NewCoalescer(sendFunc func(text string, backspaces int, method InjectionMethod)) *Coalescer {
	return &Coalescer{
		timerMs:  25,
		sendFunc: sendFunc,
	}
}

// Queue adds a replacement to the pending queue
// If a pending replacement exists, it is replaced entirely (not accumulated)
// timerMs: custom timer in ms, 0 = use default (25ms)
func (c *Coalescer) Queue(text string, backspaces int, method InjectionMethod, timerMs int) {
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

	// Use custom timer or default
	delay := c.timerMs
	if timerMs > 0 {
		delay = timerMs
	}

	c.timer = time.AfterFunc(time.Duration(delay)*time.Millisecond, func() {
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
