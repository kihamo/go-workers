package workers

import (
	"sync"
	"sync/atomic"
	"time"

	"code.cloudfoundry.org/clock"
)

type Ticker struct {
	started atomic.Value

	c      chan time.Time
	change chan time.Duration
	stop   chan struct{}

	clock  clock.Clock
	ticker clock.Ticker

	once sync.Once
}

func NewTicker(d time.Duration) *Ticker {
	c := clock.NewClock()
	t := &Ticker{
		c:      make(chan time.Time, 1),
		change: make(chan time.Duration, 1),
		stop:   make(chan struct{}, 1),
		clock:  c,
		ticker: c.NewTicker(d),
	}

	t.started.Store(false)

	return t
}

func (t *Ticker) run() {
	t.started.Store(true)

	for {
		select {
		case <-t.stop:
			if t.IsStart() {
				t.ticker.Stop()
				t.started.Store(false)
			}

			return

		case c := <-t.ticker.C():
			t.c <- c

		case d := <-t.change:
			t.ticker = t.clock.NewTicker(d)
		}
	}
}

func (t *Ticker) C() <-chan time.Time {
	t.Start()
	return t.c
}

func (t *Ticker) SetDuration(d time.Duration) {
	t.change <- d
}

func (t *Ticker) IsStart() bool {
	return t.started.Load().(bool)
}

func (t *Ticker) Start() {
	t.once.Do(func() {
		go t.run()
	})
}

func (t *Ticker) Stop() {
	t.stop <- struct{}{}
}
