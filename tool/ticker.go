package tool

import (
	"time"
)

type Ticker struct {
	ticker   *time.Ticker
	quit     chan bool
	interval time.Duration
}

func NewTicker(interval time.Duration) *Ticker {
	return &Ticker{
		interval: interval,
		quit:     make(chan bool),
		ticker:   time.NewTicker(interval),
	}

}

func (t *Ticker) Start(f func()) {
	go func() {
		for {
			select {
			case <-t.ticker.C:
				t.Do(f)
			case <-t.quit:
				t.Stop()
				return
			}
		}
	}()
}

func (t *Ticker) Stop() {
	t.quit <- true
	t.ticker.Stop()
	close(t.quit)
}
func (t *Ticker) Do(f func()) {
	f()
}
