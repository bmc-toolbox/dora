// Copyright © 2019 Joel Rebello <joel.rebello@booking.com>
// Copyright © 2019 Dmitry Verkhoturov <dmitry.verkhoturov@booking.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stats

import (
	"fmt"
	"time"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/storage"
	metrics "github.com/bmc-toolbox/gin-go-metrics"
	"github.com/spf13/viper"
)

type countable interface {
	Count(*filter.Filters) (int, error)
}

type Asset struct {
	Total         int `json:"total"`
	Updated24hAgo int `json:"updated_24h_ago"`
}

type UnitStats struct {
	Asset
	Vendors map[string]Asset `json:"by_vendor"`
}

type Stats struct {
	Uptime       float32   `json:"uptime_ms"`
	UpdateTime   string    `json:"update_time"`
	StartTime    time.Time `json:"-"`
	Chassis      UnitStats `json:"chassis"`
	Blade        UnitStats `json:"blades"`
	Discrete     UnitStats `json:"discretes"`
	Nic          UnitStats `json:"nics"`
	StorageBlade UnitStats `json:"storage_blades"`
	ScannedPort  UnitStats `json:"scanned_ports"`
	Psu          UnitStats `json:"psus"`
	Disk         UnitStats `json:"disks"`
	Fan          UnitStats `json:"fans"`
}

// UpdateUptime updates uptime based on StartTime
func (s *Stats) UpdateUptime() {
	s.Uptime = float32(time.Since(s.StartTime).Seconds() * 1e3) //1e3 == 1000
}

// GatherDBStats triggers GatherDBStats function from all resources types
func (s *Stats) GatherDBStats(
	chassisStorage *storage.ChassisStorage,
	bladeStorage *storage.BladeStorage,
	discreteStorage *storage.DiscreteStorage,
	nicStorage *storage.NicStorage,
	storageBladeStorage *storage.StorageBladeStorage,
	scannedPortStorage *storage.ScannedPortStorage,
	psuStorage *storage.PsuStorage,
	diskStorage *storage.DiskStorage,
	fanStorage *storage.FanStorage) {
	names := []string{
		"chassis",
		"blades",
		"discretes",
		"nics",
		"storage_blades",
		"scanned_ports",
		"psus",
		"disks",
		"fans"}

	for i, r := range []countable{
		chassisStorage,
		bladeStorage,
		discreteStorage,
		nicStorage,
		storageBladeStorage,
		scannedPortStorage,
		psuStorage,
		diskStorage,
		fanStorage} {
		u := &UnitStats{}
		switch i {
		case 0:
			u = &s.Chassis
		case 1:
			u = &s.Blade
		case 2:
			u = &s.Discrete
		case 3:
			u = &s.Nic
		case 4:
			u = &s.StorageBlade
		case 5:
			u = &s.ScannedPort
		case 6:
			u = &s.Psu
		case 7:
			u = &s.Disk
		case 8:
			u = &s.Fan
		}
		if u.Vendors == nil {
			u.Vendors = map[string]Asset{}
		}
		u.Total, _ = r.Count(&filter.Filters{})

		updated24hAgoFilter := &filter.Filters{}
		updated24hAgoFilter.Add("updated_at",
			[]string{"less_than", time.Now().AddDate(0, 0, -1).Format(time.RFC3339)},
			false)
		u.Updated24hAgo, _ = r.Count(updated24hAgoFilter)

		if viper.GetBool("metrics.enabled") {
			metrics.UpdateGauge([]string{fmt.Sprintf("resources.%v.total", names[i])}, int64(u.Total))
			metrics.UpdateGauge([]string{fmt.Sprintf("resources.%v.updated_24h_ago", names[i])}, int64(u.Updated24hAgo))
		}
		for _, vendor := range devices.ListSupportedVendors() {
			asset, ok := u.Vendors[vendor]
			if !ok {
				asset = Asset{}
			}
			vendorFilter := &filter.Filters{}
			vendorFilter.Add("vendor",
				[]string{vendor},
				false)
			asset.Total, _ = r.Count(vendorFilter)
			vendorFilter.Add("updated_at",
				[]string{"less_than", time.Now().AddDate(0, 0, -1).Format(time.RFC3339)},
				false)
			asset.Updated24hAgo, _ = r.Count(vendorFilter)
			u.Vendors[vendor] = asset

			if viper.GetBool("metrics.enabled") {
				metrics.UpdateGauge([]string{fmt.Sprintf("resources.%v.by_vendor.%v.total", names[i], vendor)}, int64(u.Total))
				metrics.UpdateGauge([]string{fmt.Sprintf("resources.%v.by_vendor.%v.updated_24h_ago", names[i], vendor)}, int64(u.Updated24hAgo))
			}
		}
	}
	s.UpdateTime = time.Now().Format(time.RFC3339)
}
