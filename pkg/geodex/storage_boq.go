// This file is part of silphtelescope (https://github.com/spezifisch/silphtelescope).
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
)

// BOQDB is a read-only wrapper for a Book Of Quests stops JSON.
type BOQDB struct {
	files  []string
	output chan *BOQCell
	cancel chan bool
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

// Run parses all files
func (db *BOQDB) Run() (err error) {
	run := true
	for _, file := range db.files {
		var f *os.File
		f, err = os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		br := bufio.NewReaderSize(f, 65536)
		d := json.NewDecoder(br)

		// skip the first tokens:
		// {
		// "2/123123123"
		if err = skipTokens(d, 2); err != nil {
			return
		}

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
				return
			}

			// send to output
			db.output <- &cell

			// after BOQCell skip the next token (cell id) before the next Cell starts:
			// "2/321321321"
			if err = skipTokens(d, 1); err != nil {
				return
			}
		}
		if !run {
			break
		}
	}
	return
}
