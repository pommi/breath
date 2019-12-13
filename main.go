package main

import (
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/rs/zerolog/log"
)

const (
	CONFIG_FILE_PATH = "breath.yml"
)

var (
	config Config
)

func init() {
	data, err := ioutil.ReadFile(CONFIG_FILE_PATH)
	if err != nil {
		log.Error().Msgf("Error reading file %s: %v", CONFIG_FILE_PATH, err)
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
	signal.Notify(c, os.Interrupt)
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
