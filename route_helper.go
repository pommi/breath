/**
	* The Clear BSD License
	*
	* Copyright (c) 2019 Dmitrij Igorevich
	* All rights reserved.
	*
	* Redistribution and use in source and binary forms, with or without
	*	modification, are permitted (subject to the limitations in the
	* disclaimer below) provided that the following conditions are met:
	*
	*		* Redistributions of source code must retain the above copyright notice,
	*			this list of conditions and the following disclaimer.
	*  	* Redistributions in binary form must reproduce the above copyright
	* 		notice, this list of conditions and the following disclaimer in the
	* 		documentation and/or other materials provided with the distribution.
  *		* Neither the name Dmitrij Igorevich nor the names of public
	*			contributors may be used to endorse or promote products derived from
	*			this software without specific prior written permission.
	*
	* NO EXPRESS OR IMPLIED LICENSES TO ANY PARTY'S PATENT RIGHTS ARE GRANTED BY
	* THIS LICENSE. THIS SOFTWARE IS PROVIDED BY D. IGOREVICH AND CONTRIBUTORS
	* "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING,
	* BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS
	* FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
	* HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
	* SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED
	* TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA,
	* OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY
	* OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
	* NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
	* SOFTWARE,	EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

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
