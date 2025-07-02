package scene

import "time"

type Ticker struct {
	tick int
	time time.Time
}

// Create a Ticker to synchronize two objects using ticks
func NewTicker() *Ticker {
	return &Ticker{
		tick: 0,
		time: time.Now(),
	}
}

func (t *Ticker) Reset() {
	t.tick = 0
	t.time = time.Now()
}

func (t *Ticker) Tick() (int, time.Duration) {
	t.tick++
	d := time.Since(t.time)
	t.time = time.Now()
	return t.tick, d
}

func (t *Ticker) Time() time.Time {
	return t.time
}

func (t *Ticker) Since() time.Duration {
	return time.Since(t.time)
}

func (t *Ticker) Sync(s *Ticker) int {
	t.tick = s.tick
	t.time = s.time
	return t.tick
}

func (t *Ticker) IsSynced(s *Ticker) bool {
	if t.tick == 0 || s.tick == 0 {
		return false
	}
	return t.tick == s.tick
}
