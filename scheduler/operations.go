package scheduler

type Range struct {
	Start, End int64
}

const (
	KiB int64 = 1 << 10
	MiB       = 1 << 20
	GiB       = 1 << 30
	TiB       = 1 << 40
)

// if changed, make necessary changes in `gob.go`
type mp map[int64]*Range

// properties:
// a[0] > a[1]
//
// {1,2} <- {2,3} => {1,3}
// {2,3} <- {1,2} => {1,3}
// {1,2}, {3,4} <- {2,3} => {1,4}

// merge given range to sets
func (m mp) merge(fr *Range) {
	// first set value, first set ok
	fsv, fsok := m[fr.Start]
	ssv, ssok := m[fr.End]

	if fsok && ssok {
		m.delete(fsv)
		m.delete(ssv)
		m.add(&Range{Start: fsv.Start, End: ssv.End}) // {1,2}, {3,4} ,{2,3} => {1,4}
	} else if fsok && fsv.End == fr.Start {
		m.delete(fsv)
		m.add(&Range{Start: fsv.Start, End: fr.End}) // {1,2}, {2,3} => {1,3}
	} else if ssok && ssv.Start == fr.End {
		m.delete(ssv)
		m.add(&Range{Start: fr.Start, End: ssv.End}) // {2,3}, {1,2} => {1,3}
	} else {
		m.add(fr)
	}
}

func (m mp) delete(fr *Range) {
	delete(m, fr.Start)
	delete(m, fr.End)
}

func (m mp) add(fr *Range) {
	m[fr.Start] = fr
	m[fr.End] = fr
}
