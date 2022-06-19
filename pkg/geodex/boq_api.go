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

// BOQCell are the values from the outer dict that BookOfQuests returns
type BOQCell struct {
	Stops []*BOQStop `json:"stops"`
}

// BOQStop has name and location for POIs
type BOQStop struct {
	Name      string      `json:"name"`
	IsPortal  bool        `json:"portal"`
	IsGym     bool        `json:"gym"`
	IsStop    bool        `json:"stop"`
	Timestamp int64       `json:"ts"`
	S2Level20 string      `json:"s2l20"`
	Location  BOQGeometry `json:"loc"`
}

// BOQGeometry always is a Point with Lat/Lon here
type BOQGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}
