package main

import (
  "time"
  "reflect"
  "github.com/rs/zerolog/log"
)

func (state * State) Start() {
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
      case <- state.quit:
        break
      default:
      }
    }
  }()
}

func (state * State) Cleanup() {
  state.tickers = make([]*time.Ticker, 0)
  state.groups = make([]Group, 0)
  close(state.quit)
}

func (state * State) GetChan() chan *Group {
  return state.master
}

func (state * State) UpdateAll() {
  log.Info().Msgf("Initial update of %d groups.", len(state.groups))
  for _, group := range state.groups {
    group.Update()
  }
}

func (state * State) Stop() {
  for _, t := range state.tickers {
    t.Stop()
  }
  close(state.master)
  state.quit <- struct{}{}
}
