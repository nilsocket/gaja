package scheduler

import (
	"encoding/gob"
	"log"
	"sort"
)

// gob encodes:
// type information (header), length of data, and data.
// on calling `enc.Encode()`, length of data and data is appended.
// for type of `map[int]struct{int,int}`, header length is 51 bytes
// we need only one instance of data, i.e., latest.
// we write our data from byte 51, indirectly updating data

func (s *Streaming) Write() {
	if s.pf != nil {
		if s.wroteHeader {
			s.pf.Seek(51, 0)
		} else {
			s.wroteHeader = true
		}

		err := s.enc.Encode(s.done)
		if err != nil {
			log.Println("scheduler.s.Write, enc.Encode(", s, "), err: ", err)
		}
		s.pf.Sync()
	}
}

func (s *Streaming) Load() {
	if s.pf != nil {
		err := gob.NewDecoder(s.pf).Decode(&s.done)
		if err != nil {
			log.Println("scheduler.s.Load, Decode(", s, "), err: ", err)
		}

		s.fillInfo()
	}
}

// when we resume,
// Considering a file size of 7, with piece size of 1
// currentPos = 6
// total = 3
// done = {1,2}, {4,6}
// failed = {3,4}
// in progress = {2,3}
// not scheduled = {6,7}
//
// Everything can be retrieved from `done` alone
// currentPos = max element in `done`
// total = sum of differences in sets
// failed = anything missing in `done`
// in progress = none
// not scheduled = restart from currentPos
func (s *Streaming) fillInfo() {

	ts := make([]int64, 0, len(s.done)/2)

	for k, v := range s.done {
		if k > s.currentPos {
			s.currentPos = k
		}

		// retrieve set only once into `ts`
		// retrieve only v.start from each set
		// sets doesn't overlap or equate with one another
		// it is safe to sort v.start and use it to retrieve in sorted order
		// then find missing pieces into failed
		if v.Start == k {
			ts = append(ts, k)
			s.current += (v.End - v.Start)
		}
	}

	sort.Slice(ts, func(i, j int) bool {
		return ts[i] < ts[j]
	})

	if len(ts) > 0 {
		e := s.done[ts[0]]
		if e.Start != 0 {
			s.failed.add(&Range{0, e.Start})
		}
	}

	if len(ts) > 1 {
		for i := 1; i < len(ts); i++ {
			e := s.done[ts[i-1]]
			if e.End < ts[i] {
				s.failed.add(&Range{e.End, ts[i]})
			}
		}
	}
}
