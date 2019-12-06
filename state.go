package main

import (
  "time"
  "reflect"
  "github.com/rs/zerolog/log"
)

func (state * State) Start() {
  if state.timers != nil {
    panic("Start may not be used twice")
  }

  state.timers = make([]*time.Timer, len(state.groups))
  for i, group := range state.groups {
    state.timers[i] = time.NewTimer(group.interval)
  }

  go func() {
    cases := make([]reflect.SelectCase, len(state.groups))
    for i, timer := range state.timers {
      cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(timer.C)}
    }
    index, _, _ := reflect.Select(cases)
    group := &state.groups[index]
    state.master <- group
  }()
}

func (state * State) GetChan() chan *Group {
  return state.master
}

func (state * State) Cleanup() {
  state.timers = make([]*time.Timer, 0)
  state.groups = make([]Group, 0)
}

func (state * State) UpdateAll() {
  log.Info().Msgf("Initial update of %d groups.", len(state.groups))
  for _, group := range state.groups {
    group.Update()
  }
}

func (state * State) Stop() {
  for _, timer := range state.timers {
    timer.Stop()
  }
  close(state.master)
}
