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
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/twpayne/go-kml/v2"
	"github.com/twpayne/go-kml/v2/icon"

	"github.com/spezifisch/pogo-planner/pkg/geodex"
)

type boqConverter struct {
	MapName string
	Output  io.Writer

	CellCount int
	POICount  int
	GymCount  int
	StopCount int

	gyms  []kml.Element
	stops []kml.Element
}

func (bc *boqConverter) processCell(cell *geodex.BOQCell) {
	bc.CellCount++

	for _, poi := range *cell {
		bc.POICount++
		var iconHref string
		var iconScale float64
		if poi.IsGym {
			bc.GymCount++
			// see https://kml4earth.appspot.com/icons.html
			iconHref = icon.PaddleHref("blu-stars-lv")
			iconScale = 0.7
		} else if poi.IsStop {
			bc.StopCount++
			iconHref = icon.PaddleHref("grn-circle-lv")
			iconScale = 0.5
		} else {
			continue
		}

		// check data from BOQ
		if len(poi.Location.Coordinates) != 2 {
			log.Error("invalid coordinates:", poi.Location.Coordinates)
			return
		}

		var name string
		if poi.Name != "" {
			name = poi.Name
		} else if poi.IsGym {
			name = fmt.Sprintf("Gym %d", bc.GymCount)
		} else if poi.IsStop {
			name = fmt.Sprintf("Stop %d", bc.StopCount)
		}

		fort := kml.Placemark(
			kml.Name(name),
			kml.Point(
				kml.Coordinates(kml.Coordinate{
					Lon: poi.Location.Coordinates[0],
					Lat: poi.Location.Coordinates[1],
				}),
			),
			kml.Style(
				kml.IconStyle(
					kml.Icon(
						kml.Href(
							iconHref,
						),
					),
					kml.Scale(iconScale),
				),
			),
		)
		if poi.IsGym {
			bc.gyms = append(bc.gyms, fort)
		} else if poi.IsStop {
			bc.stops = append(bc.stops, fort)
		}
	}
}

func (bc *boqConverter) generateKML() {
	wrapGymFolder := kml.Folder(
		append([]kml.Element{
			kml.Name("Gyms"),
			kml.Open(false),
		},
			bc.gyms...,
		)...,
	)

	wrapStopFolder := kml.Folder(
		append([]kml.Element{
			kml.Name("Stops"),
			kml.Open(false),
		},
			bc.stops...,
		)...,
	)

	result := kml.KML(
		kml.Document(
			kml.Name(bc.MapName),
			kml.Open(true),
			wrapGymFolder,
			wrapStopFolder,
		),
	)

	if bc.Output == nil {
		bc.Output = os.Stdout
	}
	result.WriteIndent(bc.Output, "", "  ")
}

var rootCmd = &cobra.Command{
	Use:   "genpoifile",
	Short: "Generate POI file for import in mapping applications",
	Long:  `Get Pokestop and Gym data from multiple sources and generate a GPX file.`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		outputFile, err := cmd.Flags().GetString("output")
		if err != nil || outputFile == "" {
			log.WithError(err).Error("invalid output file")
			return
		}

		output, err := os.Create(outputFile)
		if err != nil {
			log.WithField("name", outputFile).WithError(err).Error("failed creating output file")
			return
		}
		defer output.Close()

		// setup BOQ parser
		tStart := time.Now()
		boqFiles, _ := cmd.Flags().GetStringArray("boq")
		boqOutput := make(chan *geodex.BOQCell, 4)
		boqCancel := make(chan bool)
		boq, err := geodex.NewBOQDB(boqFiles, boqOutput, boqCancel)
		if err != nil {
			log.WithError(err).Error("got invalid boq files")
			return
		}

		// let BOQ reader parse all files, outputting cells to boqOutput
		go boq.Run()

		bc := boqConverter{
			MapName: fmt.Sprintf("PogoPlanner %s", time.Now().Truncate(time.Minute).String()),
			Output:  output,

			gyms:  []kml.Element{},
			stops: []kml.Element{},
		}
		count := 0
		for cell := range boqOutput {
			if cell == nil {
				break
			}
			count++
			bc.processCell(cell)
		}

		timeTrack(tStart, "boq parsing")
		tStart = time.Now()

		bc.generateKML()
		timeTrack(tStart, "kml output")

		log.Infof("processed BOQ data: %d files with %d cells containing %d POIs with %d gyms, %d stops",
			len(boqFiles), bc.CellCount, bc.POICount, bc.GymCount, bc.StopCount)
		if boq.RunError != nil {
			log.WithError(boq.RunError).Error("boq runner failed!")
		}
	},
}

// from: https://coderwall.com/p/cp5fya/measuring-execution-time-in-go
func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Infof("%s took %s", name, elapsed)
}

func main() {
	rootCmd.PersistentFlags().StringArrayP("boq", "b", []string{}, "BookOfQuests JSON file(s)")
	rootCmd.PersistentFlags().StringP("output", "o", "", "Output KML file")

	rootCmd.MarkPersistentFlagRequired("boq")
	rootCmd.MarkPersistentFlagRequired("output")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
