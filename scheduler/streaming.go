package scheduler

import (
	"encoding/gob"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/nilsocket/gaja/gfile"
	"github.com/vbauerster/mpb/v6"
)

type Streaming struct {
	size, psize, current, currentPos int64
	pf                               *os.File
	enc                              *gob.Encoder
	b                                *mpb.Bar
	wroteHeader                      bool
	done, failed                     mp
	t                                *time.Ticker
	doneChan                         chan struct{}
	*sync.Mutex
}

// NewStreaming scheduler
// bar is updated if provided
// pfp represents progress file path
// ps represents piece size
// size of the file
func NewStreaming(b *mpb.Bar, pfp string, ps, size int64) *Streaming {

	pf := gfile.OpenFiles(pfp)[0]

	t := time.NewTicker(time.Millisecond * 400)

	s := &Streaming{
		size:     size,
		psize:    getPieceSize(ps, size),
		pf:       pf,
		enc:      gob.NewEncoder(pf),
		b:        b,
		done:     make(mp),
		failed:   make(mp),
		t:        t,
		doneChan: make(chan struct{}),
		Mutex:    new(sync.Mutex),
	}

	fs, _ := pf.Stat()

	// file is resumed
	if fs.Size() > 0 {
		s.Load()
		if b != nil {
			b.SetRefill(s.current)
		}
	}

	go func() {
		for {
			select {
			case <-s.t.C:
				s.Lock()
				s.Write()
				s.Unlock()
			case <-s.doneChan:
				return
			}
		}
	}()

	return s
}

// Current position
func (s *Streaming) Current() int64 {
	s.Lock()
	defer s.Unlock()
	return s.current
}

// Range to download
// nil returned on EOF
func (s *Streaming) Range() *Range {
	for fr, done := s.next(); !done; fr, done = s.next() {
		if fr == nil {
			time.Sleep(time.Millisecond * 10)
		} else {
			return fr
		}
	}

	return nil
}

// SetSize
func (s *Streaming) SetSize(ts int64) {
	s.Lock()
	if s.b != nil {
		s.b.SetTotal(ts, false)
	}
	s.size = ts
	s.Unlock()
}

// Done updates log with given range
func (s *Streaming) Done(fr *Range) {
	s.Lock()
	s.current += (fr.End - fr.Start)

	if s.b != nil {
		s.b.SetCurrent(s.current)
	}

	s.done.merge(fr)
	s.Unlock()
}

// Failed updates log with given range
func (s *Streaming) Failed(fr *Range) {
	s.Lock()
	s.failed.merge(fr)
	s.Unlock()
}

func (s *Streaming) Close() {
	s.t.Stop()
	close(s.doneChan)
}

func (s *Streaming) next() (*Range, bool) {
	s.Lock()
	defer s.Unlock()

	if len(s.failed) > 0 {
		for _, v := range s.failed {
			s.failed.delete(v)
			return v, false
		}
	}

	prevPos := s.currentPos
	s.currentPos += s.psize

	if prevPos >= s.size {
		if len(s.done) > 2 { // some sets are being processed
			return nil, false
		} else if tfr, ok := s.done[0]; ok && tfr.End == s.size {
			// everything is done
			return nil, true
		}
		return nil, false
	} else if s.currentPos > s.size {
		s.currentPos = s.size
	}

	return &Range{Start: prevPos, End: s.currentPos}, false
}

func (fr *Range) String() string {
	startStr := strconv.FormatInt(fr.Start, 10)
	endStr := strconv.FormatInt(fr.End, 10)
	return "{" + startStr + ", " + endStr + "}"
}

func getPieceSize(ps, size int64) int64 {
	if ps == 0 {
		ps = 20 * MiB
		// if not set,
		// set it based on the size of file
		if size != 0 {
			if size <= 5*MiB {
				ps = size / 3
			} else if size <= 10*MiB {
				ps = size / 5
			} else if size <= 40*MiB {
				ps = size / 8
			} else if size <= 100*MiB {
				ps = size / 10
			}
		}
	}

	return ps
}
