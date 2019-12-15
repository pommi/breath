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
