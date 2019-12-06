package main

import (
  "github.com/rs/zerolog/log"
  "github.com/crackhd/durafmt"
  "time"
)

func (group * Group) setInterval() {

  sources := group.config.Sources[group.index]

  group.interval = time.Hour
  if len(sources.Interval) > 0 {
    duration, err := durafmt.ParseString(sources.Interval)
    if err != nil {
      log.Fatal().Msgf("sources.%d: error reading update interval string \"%s\": %v", group.index, sources.Interval, err)
    }
    group.interval = duration.Get()
  }
}

func (group * Group) resetInterval(state *State, d time.Duration) {
  state.timers[group.index].Reset(d)
  group.interval = d
}

func (group * Group) Update() {
  sources := group.config.Sources[group.index]
  log.Debug().Msgf("Updated sources.%d: %v, next update in %s", group.index, sources.Domains, durafmt.Parse(group.interval))
}
