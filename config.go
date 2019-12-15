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
