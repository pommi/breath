package main

import (
  "net"
  "github.com/rs/zerolog/log"
  "github.com/vishvananda/netlink"
)

func route_add(ip net.IP, gw *netlink.Addr, _link netlink.Link) {
	route := netlink.Route{LinkIndex: _link.Attrs().Index, Dst: gw.IPNet, Src: ip}
	if err := netlink.RouteAdd(&route); err != nil {
		log.Fatal().Msgf("route_add fail (%s, %s): %v", ip.String, gw.String(), err)
	}
}

func route_del(ip net.IP, gw *netlink.Addr, _link netlink.Link) {
	route := netlink.Route{LinkIndex: _link.Attrs().Index, Dst: gw.IPNet, Src: ip}
	if err := netlink.RouteDel(&route); err != nil {
		log.Fatal().Msgf("route_del fail (%s, %s): %v", ip.String, gw.String(), err)
	}
}

func (routeHelp * RouteHelper) Reset(linkName, gw string) {
  routeHelp.Clear()

  _link, err := netlink.LinkByName(linkName)
  if err != nil {
    log.Fatal().Msgf("RouteHelper.Reset() fail for link/iface \"%s\": %v", linkName, err)
  }

  routeHelp.link = link{
    ptr:    _link,
    name:   linkName,
  }

	routeHelp.gw, err = netlink.ParseAddr(gw)
  if err != nil {
    log.Fatal().Msgf("Failed to parse gateway address \"%s\": %v", gw, err)
  }

  routeHelp.routes = make(routesMap)
}

func (routeHelp * RouteHelper) Add(owner interface{}, _ip net.IP) {
  if routeHelp.link.ptr == nil || routeHelp.routes == nil {
    panic("RouteHelper was not initialized with an interface/gateway to use.")
  }

  ip := ipstr(_ip.String())

  if ipData, exists := routeHelp.routes[ip]; exists {
    owners := ipData.owners
    if refCount, ownerExists := owners[owner]; ownerExists {
      owners[owner] = refCount + 1
      log.Info().Msgf("ip route ADD: %v IP %s gw %s via %s (refs: %d)", owner, ip, routeHelp.gw, routeHelp.link.name, refCount + 1)
    } else {
      route_add(_ip, routeHelp.gw, routeHelp.link.ptr)
      owners[owner] = 1
      log.Info().Msgf("ip route ADD: %v IP %s gw %s via %s (refs: 1)", owner, ip, routeHelp.gw, routeHelp.link.name)
    }
  } else {
    routeHelp.routes[ip] = routeData{
      ip: _ip,
      owners: make(map[routeOwner]int),
    }
    routeHelp.routes[ip].owners[owner] = 1
    log.Info().Msgf("ip route ADD: %v IP %s gw %s via %s (refs: 1)", owner, ip, routeHelp.gw, routeHelp.link.name)
  }
}

func (routeHelp * RouteHelper) Remove(owner interface{}, _ip net.IP) {
  if routeHelp.link.ptr == nil || routeHelp.routes == nil {
    panic("RouteHelper was not initialized with an interface/gateway to use.")
  }

  ip := ipstr(_ip.String())

  if ipData, exists := routeHelp.routes[ip]; exists {
    owners := ipData.owners
    if refCount, ownerExists := owners[owner]; ownerExists {
      if refCount > 1 {
        owners[owner] = refCount - 1
      } else {
        delete(owners, owner)
        if len(owners) == 0 {
          route_del(ipData.ip, routeHelp.gw, routeHelp.link.ptr)
          delete(routeHelp.routes, ip)
        }
      }
      log.Info().Msgf("ip route REMOVE: %v IP %s gw %s via %s (remaining refs: %d)", owner, ip, routeHelp.gw, routeHelp.link.name, refCount - 1)
      return
    }
  }

  log.Warn().Msgf("ip route REMOVE fail (not found): %v IP %s gw %s via %s", owner, ip, routeHelp.gw, routeHelp.link.name)
}

func (routeHelp * RouteHelper) Clear() {
  if routeHelp.routes != nil {
    for _, ipData := range routeHelp.routes {
      route_del(ipData.ip, routeHelp.gw, routeHelp.link.ptr)
    }
    routeHelp.routes = make(routesMap)
  }
}
