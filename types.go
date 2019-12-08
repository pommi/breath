package main

import (
  "time"
  "net"
  "github.com/vishvananda/netlink"
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

type FAIL_ACTION string

const (
  ON_FAIL_HOLD FAIL_ACTION = "hold"
  ON_FAIL_DROP FAIL_ACTION = "drop"
)

type Resolver struct {
  NameServers   []string      `yaml:"nameservers,flow"`
  NameServersIP []net.IP      `yaml:"-"`
  ActionOnFail  FAIL_ACTION   `yaml:"on_failure"`
}

type State struct {
  groups      []Group
  tickers     []*time.Ticker// timeouts/intervals triggering updates for master channel
  master      chan *Group   // outer interface to listen for updates
  quit        chan struct{} // send stop signal and interrupt background loop
  routeHelp   RouteHelper
}

type Group struct {
  config * Config
  index int
  interval time.Duration
  resolver * Resolver
}

type routeOwner interface{}

type link struct {
  ptr   netlink.Link
  name  string
}

type ipstr string
type routeData struct {
  ip      net.IP
  owners  map[routeOwner]int
}
type routesMap map[ipstr]routeData

type RouteHelper struct {
  link          link          // target device
  gw            *netlink.Addr // target gateway
  routes        routesMap     // routes stored as: ip => owners
}
