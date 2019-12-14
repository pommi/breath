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
