package main

import (
  "time"
)

type ConfigChecker struct {
  Version string
}

type Config struct {
  DefaultResolver *Resolver `yaml:"default_resolver,flow"`
  Target struct {
    Name, Gateway string
  }
  Sources []struct {
    Interval string
    Domains []string `yaml:",flow"`
    Resolver *Resolver `yaml:",flow"`
  } `yaml:",flow"`
}

type Resolver struct {
  NameServers []string `yaml:"nameservers,flow"`
  ActionOnFail string `yaml:"on_fail"`

  state struct {
    onFail_HOLD bool
  } `yaml:"-"`
}

type State struct {
  groups      []Group
  tickers     []*time.Ticker// timeouts/intervals triggering updates for master channel
  master      chan *Group   // outer interface to listen for updates
  quit        chan struct{} // send stop signal and interrupt background loop
}

type Group struct {
  config * Config
  index int
  interval time.Duration
  resolver * Resolver
}

// type RouteHelper struct {
//
// }