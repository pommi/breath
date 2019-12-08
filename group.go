package main

import (
  "github.com/rs/zerolog/log"
  "github.com/crackhd/durafmt"
  "time"
)

func (group * Group) init() {

  sources := group.config.Sources[group.index]

  group.interval = time.Hour
  if len(sources.Interval) > 0 {
    duration, err := durafmt.ParseString(sources.Interval)
    if err != nil {
      log.Fatal().Msgf("sources.%d: error reading update interval string \"%s\": %v", group.index, sources.Interval, err)
    }
    group.interval = duration.Get()
  } else {
    log.Info().Msgf("sources.%d interval is not set, using 1 HOUR (\"1h\") as the default")
    group.interval = time.Hour
  }
}

func (group * Group) Update() {
  sources := group.config.Sources[group.index]
  log.Debug().Msgf("Updated sources.%d: %v, next update in %s", group.index, sources.Domains, durafmt.Parse(group.interval))
}
