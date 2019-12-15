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
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
)

const (
	// ConfigFilePath default config file name (hard-coded)
	ConfigFilePath = "breath.yml"
)

var (
	config Config
)

func init() {
	data, err := ioutil.ReadFile(ConfigFilePath)
	if err != nil {
		log.Error().Msgf("Error reading file %s: %v", ConfigFilePath, err)
		os.Exit(2)
	}

	err = LoadConfig(data, &config)
	if err != nil {
		log.Fatal().Msgf("LoadConfig() fail: %v", err)
	}
}

func main() {
	log.Info().Msg("breath starts")

	state := config.Expand()

	state.UpdateAll()

	state.Start()
	defer state.Cleanup()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range c {
			log.Warn().Msg("Interruption signal, finishing")
			state.Stop()
		}
	}()

	log.Info().Msg("Entered the loop")

	for {
		group, more := <-state.GetChan()
		if group != nil {
			group.Update(state)
		}
		if !more {
			break
		}
	}

	log.Info().Msg("Finishing (no more tasks)")
}
