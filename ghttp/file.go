package ghttp

import (
	"net/http"
	"sync"
)

// File represents HTTP File
type File struct {
	Name, Dir string
	ConcReqs  int
	Reqs      Reqs
	PreHooks  []func(f *File)
	PostHooks []func(f *File)
	SaveOpts  *SaveOpts

	getReq func() (int, *http.Request) // unexported
	putReq func(int, *http.Request)
	rwm    *sync.RWMutex
}

type Files []*File

// NewFile with single or multiple urls
func NewFile(u ...string) *File {
	return &File{
		Reqs: NewReqs(u...),
	}
}

// NewFiles with single url
func NewFiles(us ...string) Files {
	fs := make(Files, 0, len(us))
	for _, u := range us {
		fs = append(fs, NewFile(u))
	}
	return fs
}

// MultiFiles with multiple urls
func MultiFiles(uss ...[]string) Files {
	fs := make(Files, 0, len(uss))
	for _, us := range uss {
		fs = append(fs, NewFile(us...))
	}
	return fs
}

// FromReqs returns new file from given reqs
func FromReqs(rs ...*Req) *File {
	return &File{
		Reqs: rs,
	}
}

// AddURLs ,add given urls to file
func (f *File) AddURLs(us ...string) {
	f.AddReqs(NewReqs(us...)...)
}

// AddReqs to file
func (f *File) AddReqs(us ...*Req) *File {
	f.rwm.Lock()
	f.Reqs = append(f.Reqs, us...)
	f.rwm.Unlock()
	return f
}

func (fs Files) Dir(dir string) {
	for _, f := range fs {
		f.Dir = dir
	}
}

func (fs Files) PreHooks(fn ...func(*File)) {
	for _, f := range fs {
		f.PreHooks = append(f.PreHooks, fn...)
	}
}

func (fs Files) PostHooks(fn ...func(*File)) {
	for _, f := range fs {
		f.PreHooks = append(f.PreHooks, fn...)
	}
}

// DeleteURLs ,delete given urls from file
func (f *File) DeleteURLs(us ...string) {
	tm := make(map[string]struct{})

	for _, u := range us {
		tm[u] = struct{}{}
	}

	nrs := make(Reqs, 0, len(f.Reqs)-len(us))

	f.rwm.Lock()
	for _, r := range f.Reqs {
		if _, ok := tm[r.URL]; !ok {
			nrs = append(nrs, r)
		}
	}
	f.Reqs = nrs
}

// DeleteReqs , delete given reqs from file
func (f *File) DeleteReqs(rs ...*Req) {
	us := make([]string, 0, len(rs))
	for _, r := range rs {
		us = append(us, r.URL)
	}
	f.DeleteURLs(us...)
}
