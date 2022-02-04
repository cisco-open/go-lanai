package monitor

import (
	"context"
	"encoding/json"
	"strings"
)

const (
	GroupBytesAllocated DataGroup = "BytesAllocated"
	GroupGCPauses       DataGroup = "GcPauses"
	GroupCPUUsage       DataGroup = "CPUUsage"
	GroupPprof          DataGroup = "Pprof"
)

type DataGroup string

type DataStorage interface {
	Read(ctx context.Context, groups...DataGroup) (map[DataGroup]RawEntries, error)

	// Append save data entry
	Append(ctx context.Context, group DataGroup, entry interface{}, cap int64) error

	// AppendAll save all data entries, grouped by DataGroup
	AppendAll(ctx context.Context, data map[DataGroup]interface{}, cap int64) error
}

type Feed struct {
	Ts             uint64
	BytesAllocated uint64
	GcPause        uint64
	CPUUser        float64
	CPUSys         float64
	Block          int
	Goroutine      int
	Heap           int
	Mutex          int
	Threadcreate   int
}

type RawEntries []string

func (v RawEntries) MarshalJSON() (data []byte, err error) {
	if v == nil {
		return []byte("null"), nil
	}
	return []byte("[" + strings.Join(v, ",") + "]"), nil
}

type SimplePair struct {
	Ts    uint64
	Value uint64
}

func (p *SimplePair) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}

type CPUPair struct {
	Ts   uint64
	User float64
	Sys  float64
}

func (p *CPUPair) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}

type PprofPair struct {
	Ts           uint64
	Block        int
	Goroutine    int
	Heap         int
	Mutex        int
	Threadcreate int
}

func (p *PprofPair) MarshalBinary() (data []byte, err error) {
	return json.Marshal(p)
}
