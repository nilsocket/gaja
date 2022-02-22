package scheduler

import (
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
)

func TestDone1(t *testing.T) {

	st := time.Now()

	s := NewStreaming(nil, "test1.pgaja", 10, 1000000)

	// t.Logf("%#v\n", s)

	wg := &sync.WaitGroup{}

	for fr := s.Range(); fr != nil; fr = s.Range() {

		// t.Log(fr)
		wg.Add(1)
		go func(fr *Range) {
			defer wg.Done()

			if rand.Int()%7 == 0 {
				s.Failed(fr)
			} else {
				s.Done(fr)
			}

			// t.Log(fr)

		}(fr)
	}

	wg.Wait()
	s.Close()
	os.Remove("test1.pgaja")

	t.Log("Final:", s.done, s.failed, time.Since(st))
}

func TestDone2(t *testing.T) {
	st := time.Now()

	s := NewStreaming(nil, "test2.pgaja", 10, 10000)

	wg := &sync.WaitGroup{}

	wg.Add(10)
	for i := 0; i < 10; i++ {

		go func() {

			for fr := s.Range(); fr != nil; fr = s.Range() {

				if rand.Int()%7 == 0 {
					s.Failed(fr)
				} else {
					s.Done(fr)
				}
				// t.Log(fr, s)
			}
			wg.Done()
		}()
	}

	// change size in between
	s.SetSize(100000)
	s.SetSize(1000000)

	wg.Wait()
	s.Close()
	os.Remove("test2.pgaja")

	t.Log("Final:", s.done, s.failed, time.Since(st))
}
