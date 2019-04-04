// Copyright © 2019 Joel Rebello <joel.rebello@booking.com>
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

package metrics

import (
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/bmc-toolbox/dora/filter"
	"github.com/bmc-toolbox/dora/storage"
	log "github.com/sirupsen/logrus"
)

type countable interface {
	Count(*filter.Filters) (int, error)
}

type UnitStats struct {
	Total int `json:"total"`
}

type Stats struct {
	Uptime       float32   `json:"uptime_ms"`
	UpdateTime   string    `json:"update_time"`
	StartTime    time.Time `json:"-"`
	Chassis      UnitStats `json:chassis"`
	Blade        UnitStats `json:blade"`
	Discrete     UnitStats `json:discrete"`
	Nic          UnitStats `json:nic"`
	StorageBlade UnitStats `json:storage_blade"`
	ScannedPort  UnitStats `json:scanned_port"`
	Psu          UnitStats `json:psu"`
	Disk         UnitStats `json:disk"`
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
	diskStorage *storage.DiskStorage) {
	names := []string{
		"chassis",
		"blade",
		"discrete",
		"nic",
		"storage_blade",
		"scanned_port",
		"psu",
		"disk"}
	// TODO add all and updated > 24h to all resources by vendor
	for i, r := range []countable{
		chassisStorage} {
		//chassisStorage,
		//bladeStorage,
		//discreteStorage,
		//nicStorage,
		//storageBladeStorage,
		//scannedPortStorage,
		//psuStorage,
		//diskStorage} {
		s.Chassis.Total, _ = r.Count(&filter.Filters{})
		UpdateGauge([]string{fmt.Sprintf("%v.total", names[i])}, float32(s.Chassis.Total))
	}
	s.UpdateTime = time.Now().Format(time.RFC3339)
}

// Scheduler starts passed function at start and then every "interval" value
func Scheduler(interval time.Duration, fn interface{}, args ...interface{}) {
	// Set up the wrapper
	f := reflect.ValueOf(fn)
	if f.Type().NumIn() != len(args) {
		log.Errorf("incorrect number of parameters for function %v, won't be scheduled",
			runtime.FuncForPC(f.Pointer()).Name())
		return
	}
	for i := 0; i < f.Type().NumIn(); i++ {
		if f.Type().In(i) != reflect.TypeOf(args[i]) {
			log.Errorf("parameter #%v for function %v is wrong type (should be %v)",
				i,
				runtime.FuncForPC(f.Pointer()).Name(),
				f.Type().In(i))
			return
		}
	}
	inputs := make([]reflect.Value, len(args))
	for k, in := range args {
		inputs[k] = reflect.ValueOf(in)
	}
	// Run function once at interval, plus once right after start
	f.Call(inputs)
	for range time.Tick(interval) {
		f.Call(inputs)
	}
}

// GoRuntimeStats collects go runtime stats.
// prefix is a slice of metric namespace nodes.
func GoRuntimeStats(prefix []string) {

	prefix = append(prefix, "runtime")

	UpdateGauge(append(prefix, "num_goroutines"), float32(runtime.NumGoroutine()))

	var s runtime.MemStats
	runtime.ReadMemStats(&s)

	// Alloc/HeapAlloc is bytes of allocated heap objects.
	// "Allocated" heap objects include all reachable objects, as
	// well as unreachable objects that the garbage collector has
	// not yet freed. Specifically, HeapAlloc increases as heap
	// objects are allocated and decreases as the heap is swept
	// and unreachable objects are freed. Sweeping occurs
	// incrementally between GC cycles, so these two processes
	// occur simultaneously, and as a result HeapAlloc tends to
	// change smoothly (in contrast with the sawtooth that is
	// typical of stop-the-world garbage collectors).
	UpdateGauge(append(prefix, "heap_alloc"), float32(s.Alloc))

	// Sys is the total bytes of memory obtained from the OS.
	// Sys is the sum of the XSys fields below. Sys measures the
	// virtual address space reserved by the Go runtime for the
	// heap, stacks, and other internal data structures. It's
	// likely that not all of the virtual address space is backed
	// by physical memory at any given moment, though in general
	// it all was at some point.
	UpdateGauge(append(prefix, "sys"), float32(s.Sys))

	// PauseTotalNs is the cumulative nanoseconds in GC
	// stop-the-world pauses since the program started.
	//
	// During a stop-the-world pause, all goroutines are paused
	// and only the garbage collector can run.
	UpdateGauge(append(prefix, "pause_total_ns"), float32(s.PauseTotalNs))

	// NumGC is the number of completed GC cycles.
	UpdateGauge(append(prefix, "num_gc"), float32(s.NumGC))

	// HeapReleased is bytes of physical memory returned to the OS.
	//
	// This counts heap memory from idle spans that was returned
	// to the OS and has not yet been reacquired for the heap.
	UpdateGauge(append(prefix, "heap_released"), float32(s.HeapReleased))

	// HeapObjects is the number of allocated heap objects.
	//
	// Like HeapAlloc, this increases as objects are allocated and
	// decreases as the heap is swept and unreachable objects are
	// freed.
	UpdateGauge(append(prefix, "heap_objects"), float32(s.HeapReleased))
}
