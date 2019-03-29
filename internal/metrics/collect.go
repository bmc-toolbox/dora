package metrics

import (
	"runtime"
)

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
