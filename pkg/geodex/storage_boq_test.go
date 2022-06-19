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
	"reflect"
	"testing"
)

func TestBOQDB_Run(t *testing.T) {
	type fields struct {
		files  []string
		output chan *BOQCell
		cancel chan bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "no file",
			fields: fields{
				files:  []string{""},
				output: make(chan *BOQCell),
				cancel: make(chan bool),
			},
			wantErr: true,
		},
		{
			name: "non-existent file",
			fields: fields{
				files:  []string{"../../test/boq/nonexistent_foo"},
				output: make(chan *BOQCell),
				cancel: make(chan bool),
			},
			wantErr: true,
		},
		{
			name: "invalid json",
			fields: fields{
				files:  []string{"../../test/data/invalid-pokedex.json"},
				output: make(chan *BOQCell),
				cancel: make(chan bool),
			},
			wantErr: true,
		},
		{
			name: "test file",
			fields: fields{
				files:  []string{"../../test/boq/boq_stops.json"},
				output: make(chan *BOQCell, 4),
				cancel: make(chan bool),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &BOQDB{
				files:  tt.fields.files,
				output: tt.fields.output,
				cancel: tt.fields.cancel,
			}
			if err := db.Run(); (err != nil) != tt.wantErr {
				t.Errorf("BOQDB.Run() error = %v", err)
			}
		})
	}
}

func TestNewBOQDB(t *testing.T) {
	goodFiles := []string{"../../test/boq/boq_stops.json"}
	goodOutput := make(chan *BOQCell)
	goodCancel := make(chan bool)

	type args struct {
		files  []string
		output chan *BOQCell
		cancel chan bool
	}
	tests := []struct {
		name    string
		args    args
		wantDb  *BOQDB
		wantErr bool
	}{
		{
			name: "non-existent files",
			args: args{
				files:  []string{"foo", "bar", ""},
				output: goodOutput,
				cancel: goodCancel,
			},
			wantErr: true,
		},
		{
			name: "directory as file",
			args: args{
				files:  []string{"../../test/boq"},
				output: goodOutput,
				cancel: goodCancel,
			},
			wantErr: true,
		},
		{
			name: "good file",
			args: args{
				files:  goodFiles,
				output: goodOutput,
				cancel: goodCancel,
			},
			wantDb: &BOQDB{
				files:  goodFiles,
				output: goodOutput,
				cancel: goodCancel,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDb, err := NewBOQDB(tt.args.files, tt.args.output, tt.args.cancel)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBOQDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDb, tt.wantDb) {
				t.Errorf("NewBOQDB() = %v, want %v", gotDb, tt.wantDb)
			}
		})
	}
}
