// This file is part of pogo-planner (https://github.com/spezifisch/pogo-planner).
// Based on geodexgen of silphtelescope (https://github.com/spezifisch/silphtelescope).
// Copyright (C) 2021-2022 spezifisch <spezifisch-7e6@below.fr> (https://github.com/spezifisch).
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published by the
// Free Software Foundation, version 3 of the License.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License for more
// details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/spezifisch/pogo-planner/pkg/geodex"
)

var rootCmd = &cobra.Command{
	Use:   "genpoifile",
	Short: "Generate POI file for import in mapping applications",
	Long:  `Get Pokestop and Gym data from multiple sources and generate a GPX file.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		tStart := time.Now()

		// setup BOQ parser
		boqFiles, _ := cmd.Flags().GetStringArray("boq")
		boqOutput := make(chan *geodex.BOQCell)
		boqCancel := make(chan bool)
		boqDone := make(chan bool)
		boq, err := geodex.NewBOQDB(boqFiles, boqOutput, boqCancel, boqDone)
		if err != nil {
			log.WithError(err).Error("got invalid boq files")
			return
		}

		// let BOQ reader parse all files, outputting cells to boqOutput
		go boq.Run()

		boqCellCount := 0
		boqPOICount := 0
		boqGymCount := 0
		namesAdded := 0
		namesKept := 0
		for {
			done := false

			select {
			case cell := <-boqOutput:
				boqCellCount++
				for _, poi := range cell.Stops {
					boqPOICount++
					if poi.IsGym {
						boqGymCount++

						// check data from BOQ
						if len(poi.Location.Coordinates) != 2 {
							log.Error("invalid coordinates:", poi.Location.Coordinates)
							return
						}
						if poi.Name == "" {
							continue
						}

						// get gym GUID from tile db
						gymLocation := pogo.Location{
							Latitude:  poi.Location.Coordinates[1],
							Longitude: poi.Location.Coordinates[0],
						}
						tFort, err := tdb.GetNearestFort(gymLocation, 0.1)
						if err != nil {
							// fort doesn't exist in tile38 db, that's ok
							continue
						}

						// get fort from disk
						dFort, err := ddb.GetFort(*tFort.GUID)
						if err != nil {
							// doesn't exist on disk. that's ok
							continue
						}
						if dFort.Name != nil {
							// already has a name
							namesKept++
							continue
						}

						// set name and save to disk
						dFort.Name = &poi.Name
						if err = ddb.SaveFort(dFort); err != nil {
							log.Errorf("couldn't edit fort %s", *tFort.GUID)
							return
						}
						namesAdded++
					}
				}
			case <-boqDone: // boq.Run() ended
				done = true
			}

			if done {
				break
			}
		}

		timeTrack(tStart, "boq parsing")

		log.Infof("processed BOQ data: %d cells containing %d POIs with %d gyms",
			boqCellCount, boqPOICount, boqGymCount)
		log.Infof("added names to %d gyms, got %d gyms which already had a name",
			namesAdded, namesKept)
		if boq.RunError != nil {
			log.WithError(boq.RunError).Error("boq runner failed!")
		}
	},
}

// from: https://coderwall.com/p/cp5fya/measuring-execution-time-in-go
func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("> %s took %s", name, elapsed)
}

func main() {
	rootCmd.PersistentFlags().StringArrayP("boq", "b", []string{}, "BookOfQuests JSON file(s)")

	rootCmd.MarkPersistentFlagRequired("geodex")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
