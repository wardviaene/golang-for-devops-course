package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

const DATE_FORMAT = "2006-01-02T15:04:05"

type RateLimit struct {
	mu            sync.Mutex
	hits          map[string]uint64
	limitExceeded bool
	limitLifted   time.Time
}

func (rl *RateLimit) ratelimit(w http.ResponseWriter, r *http.Request) {
	if rl.limitExceeded && time.Now().Before(rl.limitLifted) {
		w.WriteHeader(429)
		w.Write([]byte("Rate Limited"))
		return
	}
	if rl.limitExceeded && time.Now().After(rl.limitLifted) {
		rl.limitExceeded = false
	}
	rl.mu.Lock()
	timestamp := time.Now()
	strTimestamp := timestamp.Format(DATE_FORMAT)
	if val, ok := rl.hits[strTimestamp]; ok {
		if val == 5 {
			rl.limitExceeded = true
			rl.limitLifted = time.Now().Add(time.Second * 10)
		} else {
			rl.hits[strTimestamp] = val + 1
		}
	} else {
		rl.hits[strTimestamp] = 1
	}
	rl.mu.Unlock()
	timestampOneSecondEarlier := time.Now().Add(time.Duration(-1) * time.Second)
	if rl.hits[timestampOneSecondEarlier.Format(DATE_FORMAT)] == 5 {
		w.Write([]byte(fmt.Sprintf("DONE! You did it! Hitting API at %d requests in a given second\n", rl.hits[timestampOneSecondEarlier.Format(DATE_FORMAT)])))
	} else {
		w.Write([]byte(fmt.Sprintf("Hitting API at %d requests in a given second\n", rl.hits[timestampOneSecondEarlier.Format(DATE_FORMAT)])))
	}
	if len(rl.hits) > 100000 {
		rl.mu.Lock()
		fmt.Printf("Map is getting big, resetting...\n")
		oldVal := rl.hits[strTimestamp]
		rl.hits = make(map[string]uint64)
		rl.hits[strTimestamp] = oldVal
		time.Sleep(1 * time.Second)
		rl.mu.Unlock()
	}
}
