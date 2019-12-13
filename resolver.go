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
