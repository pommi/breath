package main

import (
	"net"
	"time"

	"github.com/crackhd/durafmt"
	"github.com/rs/zerolog/log"
)

func (group *Group) init() {

	sources := group.config.Sources[group.index]

	group.interval = time.Hour
	if len(sources.Interval) > 0 {
		duration, err := durafmt.ParseString(sources.Interval)
		if err != nil {
			log.Fatal().Msgf("sources.%d: error reading update interval string \"%s\": %v", group.index, sources.Interval, err)
		}
		group.interval = duration.Get()
	} else {
		log.Info().Msgf("sources.%d interval is not set, using 1 HOUR (\"1h\") as the default", group.index)
		group.interval = time.Hour
	}
}

// Update group by adding and removing routed IPs using group domain list and resolver
func (group *Group) Update(state *State) {
	sources := group.config.Sources[group.index]
	log.Debug().Msgf("Updating sources.%d (%d domains) (DNS: %v)", group.index, len(sources.Domains), group.resolver.NameServersIP)

	routedIPs := make([]net.IP, 0)
	for _, domain := range sources.Domains {
		log.Debug().Msgf("RESOLVE: %s", domain)
		ips, err := group.resolver.Resolve(domain)
		if err != nil {
			log.Warn().Msgf("sources.%d RESOLOVE FAIL for domain: %s: %v (skipping)", group.index, domain, err)
			// TODO: support on_failure: "hold"
		} else {
			log.Debug().Msgf("%s: %v", domain, ips)
			routedIPs = append(routedIPs, ips...)
		}
	}

	state.helper.Replace(group.index, routedIPs)

	log.Debug().Msgf("Updated sources.%d (%d domains), next update in %s", group.index, len(sources.Domains), durafmt.Parse(group.interval))
}
