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
	"reflect"
	"time"

	"github.com/rs/zerolog/log"
)

// Start tickers and polling from group timers, so that the [master] channel
// will receive tasks
func (state *State) Start() {
	if state.tickers != nil {
		panic("Start may not be used twice")
	}

	state.tickers = make([]*time.Ticker, len(state.groups))
	for i, group := range state.groups {
		state.tickers[i] = time.NewTicker(group.interval)
	}

	go func() {
		cases := make([]reflect.SelectCase, len(state.groups))
		for i, t := range state.tickers {
			cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(t.C)}
		}

		for {
			index, _, _ := reflect.Select(cases)
			group := &state.groups[index]
			state.master <- group

			select {
			case <-state.quit:
				break
			default:
			}
		}
	}()
}

// Cleanup disposes of any resource or goroutine created internaly by State
func (state *State) Cleanup() {
	state.helper.Flush()
	state.tickers = make([]*time.Ticker, 0)
	state.groups = make([]Group, 0)
	close(state.quit)
}

// GetChan to use as task output channel (receive groups to update in time)
func (state *State) GetChan() chan *Group {
	return state.master
}

// UpdateAll performs out-of-order update of each source group
func (state *State) UpdateAll() {
	log.Info().Msgf("Initial update of %d groups.", len(state.groups))
	for _, group := range state.groups {
		group.Update(state)
	}
}

// Stop to interrupt channel, stop all tickers and further tasks
func (state *State) Stop() {
	for _, t := range state.tickers {
		t.Stop()
	}
	close(state.master)
	state.quit <- struct{}{}
}
