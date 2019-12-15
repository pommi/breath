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
	"errors"
	"fmt"
	"net"

	dns_impl "github.com/miekg/dns"
	"github.com/rs/zerolog/log"
	_ "github.com/vishvananda/netlink"
)

func resolve(target string, dns net.IP) ([]net.IP, error) {
	server := dns.String()

	c := dns_impl.Client{}
	m := dns_impl.Msg{}
	m.SetQuestion(target+".", dns_impl.TypeA)
	r, _, err := c.Exchange(&m, server+":53")
	if err != nil {
		return nil, err
	}

	results := make([]net.IP, len(r.Answer))

	for i, ans := range r.Answer {
		Arecord := ans.(*dns_impl.A)
		results[i] = Arecord.A
	}

	return results, nil
}

func (resolver *Resolver) init() error {

	if len(resolver.ActionOnFail) == 0 {
		resolver.ActionOnFail = FailActionDROP
		log.Info().Msgf("When on_failure is not specified, \"%s\" will be effective action.", resolver.ActionOnFail)
	} else if resolver.ActionOnFail != FailActionDROP && resolver.ActionOnFail != FailActionHOLD {
		msg := fmt.Sprintf("unsupported value \"%s\" for option \"on_failure\"", resolver.ActionOnFail)
		return errors.New(msg)
	}

	if len(resolver.NameServers) == 0 {
		return errors.New("No nameservers specified")
	}

	resolver.NameServersIP = make([]net.IP, len(resolver.NameServers))
	for i, dns := range resolver.NameServers {
		ip := net.ParseIP(dns)
		if ip == nil {
			msg := fmt.Sprintf("Nameserver \"%s\" is not valid IP address", dns)
			return errors.New(msg)
		}
		resolver.NameServersIP[i] = ip
	}

	return nil

}

// Resolve to get all domain name A records
func (resolver *Resolver) Resolve(domain string) ([]net.IP, error) {
	var (
		result []net.IP
		err    error
	)

	for i, dns := range resolver.NameServersIP {
		result, err = resolve(domain, dns)
		if err == nil {
			break
		}
		log.Warn().Msgf("Resolution failed using DNS %s domain %s type A: %v (%d/%d)", resolver.NameServers[i], domain, err, i+1, len(resolver.NameServersIP))
	}

	return result, err
}
