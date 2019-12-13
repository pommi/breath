package main

import (
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
	"github.com/vishvananda/netlink"
)

// Add gateway route for single IP address
func route_add(ip *net.IPNet, gw net.IP, link netlink.Link) {
	log.Info().Msgf("route_add: %s via %s onlink", ip.IP, gw)
	route := netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       ip,
		Gw:        gw,
		Priority:  600,
		Flags:     int(netlink.FLAG_ONLINK),
	}
	if err := netlink.RouteAdd(&route); err != nil {
		log.Error().Msgf("route_add fail (%s, %s): %v", ip.String(), gw.String(), err)
	}
}

// Delete gateway route for single IP address
func route_del(ip *net.IPNet, gw net.IP, link netlink.Link) {
	log.Info().Msgf("route_del: %s via %s onlink", ip.IP, gw)
	route := netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       ip,
		Gw:        gw,
		Priority:  600,
		Flags:     int(netlink.FLAG_ONLINK),
	}
	if err := netlink.RouteDel(&route); err != nil {
		log.Error().Msgf("route_del fail (%s, %s): %v", ip.String(), gw.String(), err)
	}
}

// Reset helper for use with new link and target gateway IP
func (routeHelp *RouteHelper) Reset(linkName, gw string) {
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

// Tell link (interface) name from internal pointer
func (routeHelp *RouteHelper) linkName() string {
	if routeHelp.link != nil {
		if as := routeHelp.link.Attrs(); as != nil {
			if len(as.Name) > 0 {
				return as.Name
			}
		}
	}

	return "link-noname"
}

// Add route (phusically, if new) with ownership and
// option to avoid duplication (othwerise, increase refcount of the route)
func (routeHelp *RouteHelper) Add(owner interface{}, _ip net.IP, increaseRef bool) {
	if routeHelp.link == nil || routeHelp.routes == nil {
		panic("RouteHelper was not initialized with an interface/gateway to use.")
	}

	ip := ipstr(_ip.String())
	dst := &net.IPNet{
		IP:   _ip,
		Mask: net.CIDRMask(32, 32),
	}

	if ipData, exists := routeHelp.routes[ip]; exists {
		owners := ipData.owners
		if refCount, ownerExists := owners[owner]; ownerExists {
			log.Warn().Msg("duplicated add")
			if increaseRef {
				owners[owner] = refCount + 1
			}
		} else {
			owners[owner] = 1
		}
	} else {
		routeHelp.routes[ip] = routeData{
			ip:     dst,
			owners: make(map[routeOwner]int),
		}
		routeHelp.routes[ip].owners[owner] = 1
		route_add(dst, routeHelp.gw, routeHelp.link)
	}
}

// Remove single reference to a route. If there are no more owners
// and references to it, route is deleted physically.
func (routeHelp *RouteHelper) Remove(owner interface{}, _ip net.IP) int {
	if routeHelp.link == nil || routeHelp.routes == nil {
		panic("RouteHelper was not initialized with an interface/gateway to use.")
	}

	ip := ipstr(_ip.String())
	dst := &net.IPNet{
		IP:   _ip,
		Mask: net.CIDRMask(32, 32),
	}

	if ipData, exists := routeHelp.routes[ip]; exists {
		owners := ipData.owners
		if refCount, ownerExists := owners[owner]; ownerExists {
			if refCount > 1 {
				owners[owner] = refCount - 1
				return refCount - 1
			} else {
				delete(owners, owner)
				if len(owners) == 0 {
					route_del(dst, routeHelp.gw, routeHelp.link)
					delete(routeHelp.routes, ip)
				}
			}
			return 0
		}
	}

	return -1
}

// Clear destroys all physical routes set up by this helper
func (routeHelp *RouteHelper) Clear() {
	if routeHelp.routes != nil {
		log.Warn().Msg("CLEAR: Performing DELETE on all added routes")
		for _, ipData := range routeHelp.routes {
			route_del(ipData.ip, routeHelp.gw, routeHelp.link)
		}
		routeHelp.routes = make(routesMap)
	}
}

// Replace adds multiple routes. Erase all previous routes by this owner.
// Change reference count to 1 for owner routes.
func (routeHelp *RouteHelper) Replace(owner interface{}, ips []net.IP) {
	ipsToRemove := make([]net.IP, 0)

	for _, ip := range ips {
		routeHelp.Add(owner, ip, false)
	}

	for _, ipData := range routeHelp.routes {
		owners := ipData.owners
		if _, ownerExists := owners[owner]; ownerExists {
			if containsIP(ips, ipData.ip.IP) {
				owners[owner] = 1
			} else {
				ipsToRemove = append(ipsToRemove, ipData.ip.IP)
			}
		}
	}

	for _, ip := range ipsToRemove {
		routeHelp.Remove(owner, ip)
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
