// This file is part of pogo-planner (https://github.com/spezifisch/pogo-planner).
// Based on silphtelescope (https://github.com/spezifisch/silphtelescope).
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

package geodex

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// BOQDB is a read-only wrapper for a Book Of Quests stops JSON.
type BOQDB struct {
	RunError error
	files    []string
	output   chan *BOQCell
	cancel   chan bool
}

// NewBOQDB returns a ready-to-use BOQDB object
func NewBOQDB(files []string, output chan *BOQCell, cancel chan bool) (db *BOQDB, err error) {
	err = checkFiles(files)
	if err != nil {
		return
	}

	return &BOQDB{
		files:  files,
		output: output,
		cancel: cancel,
	}, nil
}

func checkFiles(files []string) (err error) {
	for _, file := range files {
		var fi os.FileInfo
		fi, err = os.Stat(file)
		if err != nil {
			return
		}

		if !fi.Mode().IsRegular() {
			text := fmt.Sprintf("'%s' is not a file", file)
			return errors.New(text)
		}
	}
	return
}

func skipTokens(d *json.Decoder, count int) (err error) {
	// skip $count tokens
	for i := 0; i < count; i++ {
		_, err = d.Token()
		if err != nil {
			return
		}
	}
	return
}

func (db *BOQDB) signalDone() {
	log.Info("boq parser done signal")
	db.output <- nil
}

// Run parses all files
func (db *BOQDB) Run() (err error) {
	db.RunError = nil
	defer db.signalDone()
	run := true
	log.WithField("files", db.files).Info("starting boq parser")
	for _, file := range db.files {
		var f *os.File
		f, err = os.Open(file)
		if err != nil {
			db.RunError = err
			return
		}
		defer f.Close()

		br := bufio.NewReaderSize(f, 65536)
		d := json.NewDecoder(br)

		for d.More() {
			// check for cancel signal
			select {
			case <-db.cancel:
				run = false
			default:
			}
			if !run {
				break
			}

			// json decode cell
			var cell BOQCell
			err = d.Decode(&cell)
			if err != nil {
				db.RunError = err
				log.WithError(err).Error("cell decode failed")
				return
			}

			// send to output
			db.output <- &cell
		}
		if !run {
			break
		}
	}
	log.Info("boq parser returns ok")
	return
}
