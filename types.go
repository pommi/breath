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
	"net"
	"time"

	"github.com/vishvananda/netlink"
)

// ConfigChecker is usable for YAML file validation before loading Config
type ConfigChecker struct {
	Version string
}

// Config is an input data layout
type Config struct {
	DefaultResolver *Resolver `yaml:"default_resolver,flow"`
	Target          struct {
		Name, Gateway string
		Metric        int
	}
	Sources []struct {
		Interval string
		Domains  []string  `yaml:",flow"`
		Resolver *Resolver `yaml:",flow"`
	} `yaml:",flow"`
}

// FailAction support is not ready (TODO)
type FailAction string

const (
	// FailActionDROP resolution error will cause dropped route, which is risky
	FailActionDROP FailAction = "drop"
	// FailActionHOLD should be implemented to allow persisting routes
	// until resolution reports any other IP address or addresses for the domain.
	FailActionHOLD FailAction = "hold"
)

// Resolver performs DNS resolution with options. Each group can use
// default_resolver, or define its own resolver.
type Resolver struct {
	NameServers   []string   `yaml:"nameservers,flow"`
	NameServersIP []net.IP   `yaml:"-"`
	ActionOnFail  FailAction `yaml:"on_failure"`
}

// State is an expanded configuration
type State struct {
	groups  []Group
	tickers []*time.Ticker // timeouts/intervals triggering updates for master channel
	master  chan *Group    // outer interface to listen for updates
	quit    chan struct{}  // send stop signal and interrupt background loop
	helper  RouteHelper
}

// GroupID is an index of group, used as an identifier
type GroupID int

// Group is list of domain names with Resolver attached to it, used
// to generate and maintain IP routes
type Group struct {
	config   *Config
	index    GroupID
	interval time.Duration
	resolver *Resolver
}

type ipstr string
type routeData struct {
	dst    *net.IPNet
	owners map[GroupID]int
}
type routesMap map[ipstr]routeData

// RouteHelper is used to maintain routes from multiple groups with possible IP intersections
// still gives a way to track reference count for each
type RouteHelper struct {
	link   netlink.Link // target device
	gw     net.IP       // target gateway
	metric int          // route metric
	routes routesMap    // routes stored as: ip => owners
}
