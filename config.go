package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// LoadConfig parse config file
func LoadConfig(data []byte, config *Config) error {
	var (
		check ConfigChecker
		err   error
	)

	err = yaml.Unmarshal([]byte(data), &check)
	if err != nil {
		msg := fmt.Sprintf("YAML.Unmarshal error (checking version): %v", err)
		return errors.New(msg)
	}

	if check.Version != "1" {
		msg := fmt.Sprintf("Version \"%s\" in config file is not supported.", check.Version)
		return errors.New(msg)
	}

	err = yaml.Unmarshal([]byte(data), config)
	if err != nil {
		msg := fmt.Sprintf("YAML.Unmarshal error (loading config): %v", err)
		return errors.New(msg)
	}

	return nil
}

// Expand Config to State
func (config *Config) Expand() *State {
	groups := make([]Group, len(config.Sources))
	if len(groups) == 0 {
		log.Fatal().Msg("Config does not have any sources/groups.")
	}

	if config.DefaultResolver == nil {
		log.Fatal().Msg("default_resolver must be specified")
	}

	if len(config.Target.Name) == 0 || strings.Contains(config.Target.Name, " ") {
		log.Fatal().Msgf("Invalid target.name (interface/link) \"%s\"", config.Target.Name)
	}
	if len(config.Target.Gateway) == 0 || strings.Contains(config.Target.Gateway, "/") {
		log.Fatal().Msgf("Invalid target.gateway (IP): %s", config.Target.Gateway)
	}

	err := config.DefaultResolver.init()
	if err != nil {
		log.Fatal().Msgf("default_resolver init fail: %v", err)
	}

	for i, sources := range config.Sources {
		if sources.Resolver != nil {
			err = sources.Resolver.init()
			if err != nil {
				log.Fatal().Msgf("sources.%d resolver init fail: %v", i, err)
			}
		}
	}

	for i := range groups {
		groups[i].index = GroupID(i)
		groups[i].config = config

		groups[i].init()

		groups[i].resolver = config.Sources[i].Resolver
		if groups[i].resolver == nil {
			groups[i].resolver = config.DefaultResolver
		}

	}

	state := &State{
		groups:  groups,
		tickers: nil,
		master:  make(chan *Group),
		quit:    make(chan struct{}),
	}

	state.helper.Reset(config.Target.Name, config.Target.Gateway, config.Target.Metric)

	return state
}
