package main

import (
	"reflect"
	"time"

	"github.com/rs/zerolog/log"
)

// Start tickers and polling from group timers, so that the [master] channel
// will receive tasks
func (state *State) Start() {
	if state.tickers != nil {
		panic("Start may not be used twice")
	}

	state.tickers = make([]*time.Ticker, len(state.groups))
	for i, group := range state.groups {
		state.tickers[i] = time.NewTicker(group.interval)
	}

	go func() {
		cases := make([]reflect.SelectCase, len(state.groups))
		for i, t := range state.tickers {
			cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(t.C)}
		}

		for {
			index, _, _ := reflect.Select(cases)
			group := &state.groups[index]
			state.master <- group

			select {
			case <-state.quit:
				break
			default:
			}
		}
	}()
}

// Cleanup disposes of any resource or goroutine created internaly by State
func (state *State) Cleanup() {
	state.helper.Flush()
	state.tickers = make([]*time.Ticker, 0)
	state.groups = make([]Group, 0)
	close(state.quit)
}

// GetChan to use as task output channel (receive groups to update in time)
func (state *State) GetChan() chan *Group {
	return state.master
}

// UpdateAll performs out-of-order update of each source group
func (state *State) UpdateAll() {
	log.Info().Msgf("Initial update of %d groups.", len(state.groups))
	for _, group := range state.groups {
		group.Update(state)
	}
}

// Stop to interrupt channel, stop all tickers and further tasks
func (state *State) Stop() {
	for _, t := range state.tickers {
		t.Stop()
	}
	close(state.master)
	state.quit <- struct{}{}
}
