package main

import (
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
	"github.com/vishvananda/netlink"
)

func (helper *RouteHelper) mkRoute(ip *net.IPNet, gw net.IP, link netlink.Link) netlink.Route {
	return netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       ip,
		Gw:        gw,
		Priority:  helper.metric,
		Flags:     int(netlink.FLAG_ONLINK),
	}
}

func (helper *RouteHelper) addRoute(ip *net.IPNet, gw net.IP, link netlink.Link) {
	log.Info().Msgf("ROUTE ADD: %s via %s dev %s onlink", ip, gw, helper.linkName())
	route := helper.mkRoute(ip, gw, link)
	if err := netlink.RouteAdd(&route); err != nil {
		log.Error().Msgf("route_add fail (%s, %s, %s): %v", ip.String(), gw.String(), helper.linkName(), err)
	}
}

func (helper *RouteHelper) rmRoute(ip *net.IPNet, gw net.IP, link netlink.Link) {
	log.Info().Msgf("ROUTE DEL: %s via %s dev %s onlink", ip, gw, helper.linkName())
	route := helper.mkRoute(ip, gw, link)
	if err := netlink.RouteDel(&route); err != nil {
		log.Error().Msgf("route_del fail (%s, %s): %v", ip.String(), gw.String(), err)
	}
}

// Reset helper for use with new link and target gateway IP
func (helper *RouteHelper) Reset(linkName, gw string, metric int) {
	helper.Flush()

	var err error

	helper.link, err = netlink.LinkByName(linkName)
	if err != nil {
		msg := fmt.Sprintf("RouteHelper.Reset() fail for link/iface \"%s\": %v", linkName, err)
		if err == netlink.ErrNotImplemented {
			msg += ". Netlink library reported no-support for effective environment or operating system."
		}
		log.Fatal().Msg(msg)
	}

	helper.gw = net.ParseIP(gw)
	if helper.gw == nil {
		log.Fatal().Msgf("Failed to parse gateway address \"%s\"", gw)
	}

	helper.metric = metric

	helper.routes = make(routesMap)
}

// Tell link (interface) name from internal pointer
func (helper *RouteHelper) linkName() string {
	if helper.link != nil {
		if as := helper.link.Attrs(); as != nil {
			if len(as.Name) > 0 {
				return as.Name
			}
		}
	}

	return "link-noname"
}

// Add route (phusically, if new) with ownership and
// option to avoid duplication (othwerise, increase refcount of the route)
func (helper *RouteHelper) Add(owner GroupID, ip net.IP, increaseRef bool) {
	if helper.link == nil || helper.routes == nil {
		panic("RouteHelper was not initialized with an interface/gateway to use.")
	}

	key := ipstr(ip.String())

	if ipData, exists := helper.routes[key]; exists {
		owners := ipData.owners
		if refCount, ownerExists := owners[owner]; ownerExists {
			if increaseRef {
				owners[owner] = refCount + 1
			}
		} else {
			owners[owner] = 1
		}
	} else {
		dst := &net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(32, 32),
		}
		helper.routes[key] = routeData{
			dst:    dst,
			owners: make(map[GroupID]int),
		}
		helper.routes[key].owners[owner] = 1
		helper.addRoute(dst, helper.gw, helper.link)
	}
}

// Remove single reference to a route. If there are no more owners
// and references to it, route is deleted physically.
func (helper *RouteHelper) Remove(owner GroupID, ip net.IP) int {
	if helper.link == nil || helper.routes == nil {
		panic("RouteHelper was not initialized with an interface/gateway to use.")
	}

	key := ipstr(ip.String())

	if ipData, exists := helper.routes[key]; exists {
		owners := ipData.owners
		if refCount, ownerExists := owners[owner]; ownerExists {
			if refCount > 1 {
				owners[owner] = refCount - 1
				return refCount - 1
			}

			delete(owners, owner)
			if len(owners) == 0 {
				helper.rmRoute(ipData.dst, helper.gw, helper.link)
				delete(helper.routes, key)
			}

			return 0
		}
	}

	return -1
}

// Flush to destroy all physical routes set up by this helper
func (helper *RouteHelper) Flush() {
	if helper.routes != nil {
		log.Warn().Msg("CLEAR: Performing DELETE on all added routes")
		for _, ipData := range helper.routes {
			helper.rmRoute(ipData.dst, helper.gw, helper.link)
		}
		helper.routes = make(routesMap)
	}
}

// Replace adds multiple routes. Erase all previous routes by this owner.
// Change reference count to 1 for owner routes.
func (helper *RouteHelper) Replace(owner GroupID, ips []net.IP) {

	for _, ip := range ips {
		helper.Add(owner, ip, false)
	}

	for key, ipData := range helper.routes {
		owners := ipData.owners
		if _, ownerExists := owners[owner]; ownerExists {
			if !containsIP(ips, ipData.dst.IP) {
				delete(owners, owner)
			}
		}

		if len(owners) == 0 {
			delete(helper.routes, key)
			helper.rmRoute(ipData.dst, helper.gw, helper.link)
		}
	}
}

func containsIP(ips []net.IP, ip net.IP) bool {
	for _, _ip := range ips {
		if ip.Equal(_ip) {
			return true
		}
	}
	return false
}
