package main

import (
  "net"
  "fmt"
  "github.com/rs/zerolog/log"
  "github.com/vishvananda/netlink"
)

func route_add(ip *net.IPNet, gw net.IP, link netlink.Link) {
	route := netlink.Route{
    LinkIndex:  link.Attrs().Index,
    Dst:        ip,
    Gw:         gw,
    Priority:   600,
  }
	if err := netlink.RouteAdd(&route); err != nil {
		log.Fatal().Msgf("route_add fail (%s, %s): %v", ip.String(), gw.String(), err)
	}
}

func route_del(ip *net.IPNet, gw net.IP, link netlink.Link) {
	route := netlink.Route{
    LinkIndex:  link.Attrs().Index,
    Dst:        ip,
    Gw:         gw,
    Priority:   600,
  }
	if err := netlink.RouteDel(&route); err != nil {
		log.Fatal().Msgf("route_del fail (%s, %s): %v", ip.String(), gw.String(), err)
	}
}

func (routeHelp * RouteHelper) Reset(linkName, gw string) {
  routeHelp.Clear()

  var err error

  routeHelp.link, err = netlink.LinkByName(linkName)
  if err != nil {
    msg := fmt.Sprintf("RouteHelper.Reset() fail for link/iface \"%s\": %v", linkName, err)
    if err == netlink.ErrNotImplemented {
      msg += ". Netlink library reported no-support for effective environment or operating system."
    }
    log.Fatal().Msg(msg)
  }

	routeHelp.gw = net.ParseIP(gw)
  if routeHelp.gw == nil {
    log.Fatal().Msgf("Failed to parse gateway address \"%s\"", gw)
  }

  routeHelp.routes = make(routesMap)
}

func (routeHelp * RouteHelper) linkName() string {
  if routeHelp.link != nil {
    if as := routeHelp.link.Attrs(); as != nil {
      if len(as.Name) > 0 {
        return as.Name
      }
    }
  }

  return "link-noname"
}

func (routeHelp * RouteHelper) Add(owner interface{}, _ip net.IP, duplicateRef bool) {
  if routeHelp.link == nil || routeHelp.routes == nil {
    panic("RouteHelper was not initialized with an interface/gateway to use.")
  }

  ip := ipstr(_ip.String())
  dst := &net.IPNet{
		IP:   _ip,
		Mask: net.CIDRMask(24, 32),
	}

  if ipData, exists := routeHelp.routes[ip]; exists {
    owners := ipData.owners
    if refCount, ownerExists := owners[owner]; ownerExists {
      if duplicateRef {
        owners[owner] = refCount + 1
      }
      log.Debug().Msgf("ip route ADD: %v IP %s gw %s via %s (refs: %d)", owner, ip, routeHelp.gw, routeHelp.linkName(), refCount + 1)
    } else {
      route_add(dst, routeHelp.gw, routeHelp.link)
      owners[owner] = 1
      log.Info().Msgf("ip route ADD: %v IP %s gw %s via %s (refs: 1)", owner, ip, routeHelp.gw, routeHelp.linkName())
    }
  } else {
    routeHelp.routes[ip] = routeData{
      ip:     dst,
      owners: make(map[routeOwner]int),
    }
    routeHelp.routes[ip].owners[owner] = 1
    log.Info().Msgf("ip route ADD: %v IP %s gw %s via %s (refs: 1)", owner, ip, routeHelp.gw, routeHelp.linkName())
  }
}

func (routeHelp * RouteHelper) Remove(owner interface{}, _ip net.IP) int {
  if routeHelp.link == nil || routeHelp.routes == nil {
    panic("RouteHelper was not initialized with an interface/gateway to use.")
  }

  ip := ipstr(_ip.String())
  dst := &net.IPNet{
		IP:   _ip,
		Mask: net.CIDRMask(24, 32),
	}

  if ipData, exists := routeHelp.routes[ip]; exists {
    owners := ipData.owners
    if refCount, ownerExists := owners[owner]; ownerExists {
      if refCount > 1 {
        owners[owner] = refCount - 1
        log.Debug().Msgf("ip route REMOVE: %v IP %s gw %s via %s (refs: %d)", owner, ip, routeHelp.gw, routeHelp.linkName(), refCount - 1)
        return refCount - 1
      } else {
        delete(owners, owner)
        if len(owners) == 0 {
          route_del(dst, routeHelp.gw, routeHelp.link)
          delete(routeHelp.routes, ip)
        }
        log.Info().Msgf("ip route REMOVE: %v IP %s gw %s via %s (owners: %d)", owner, ip, routeHelp.gw, routeHelp.linkName(), len(owners))
      }
      return 0
    }
  }

  log.Warn().Msgf("ip route REMOVE fail: %v IP %s gw %s via %s (not found)", owner, ip, routeHelp.gw, routeHelp.linkName())
  return -1
}

func (routeHelp * RouteHelper) Clear() {
  if routeHelp.routes != nil {
    for _, ipData := range routeHelp.routes {
      route_del(ipData.ip, routeHelp.gw, routeHelp.link)
    }
    routeHelp.routes = make(routesMap)
  }
}
