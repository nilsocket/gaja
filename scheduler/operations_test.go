package scheduler

import (
	"errors"
	"strconv"
	"testing"
)

func TestMerge(t *testing.T) {
	if err := testMerge(10, 100000); err != nil {
		t.Error(err)
	}
}

func BenchmarkMergeTenThousand(b *testing.B)    { benchmarkMerge(10, 10000, b) }
func BenchmarkMergeMillion(b *testing.B)        { benchmarkMerge(10, 1000000, b) }
func BenchmarkMergeTenMillion(b *testing.B)     { benchmarkMerge(10, 10000000, b) }
func BenchmarkMergeHundredMillion(b *testing.B) { benchmarkMerge(10, 100000000, b) }

func benchmarkMerge(psize, size int64, b *testing.B) {
	for n := 0; n < b.N; n++ {
		testMerge(psize, size)
	}
}

func testMerge(psize, size int64) error {
	m := make(mp)
	rs := generateRanges(psize, size)

	for _, r := range rs {
		m.merge(r)
	}

	return isMapDone(m, size)
}

func isMapDone(m mp, size int64) error {
	if len(m) == 2 {
		r := m[0]
		if r.End != size {
			return errors.New("Merge not complete:" + r.String())
		}
		return nil
	}

	return errors.New("len of map," + strconv.Itoa(len(m)) + " != 2")
}

func generateRanges(psize, size int64) []*Range {
	var i int64

	ranges := make([]*Range, 0, (size/psize)+1)

	for i = 0; i < size; i += psize {
		ranges = append(ranges, &Range{i, i + psize})
	}

	return ranges
}
