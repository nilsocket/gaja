package ghttp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/nilsocket/gaja/gfile"
)

// Req represents HTTP Request
type Req struct {
	URL                string
	cookies            []*http.Cookie
	header             http.Header
	username, password string
	p                  *sync.Pool
}

// NewReq from url
func NewReq(u string) *Req {
	return &Req{
		URL:    u,
		header: make(http.Header),
	}
}

func (r *Req) SetCookies(cs []*http.Cookie) *Req {
	r.cookies = cs
	return r
}

func (r *Req) Cookies() []*http.Cookie {
	return r.cookies
}

func (r *Req) SetHeader(h http.Header) *Req {
	r.header = h
	return r
}

func (r *Req) Header() http.Header {
	return r.header
}

func (r *Req) SetUserAgent(ua string) *Req {
	r.header.Set("User-Agent", ua)
	return r
}

func (r *Req) SetReferer(ref string) *Req {
	r.header.Set("Referer", ref)
	return r
}

func (r *Req) SetAuth(username, password string) *Req {
	r.username = username
	r.password = password
	return r
}

func (r *Req) ToRequest(method string) *http.Request {
	req, err := http.NewRequest(method, r.URL, nil)
	if err != nil {
		log.Println("ghttp.r.newRequest, http.NewRequest(", r.URL, ") err: ", err)
	}

	req.Header = r.header

	if r.username != "" {
		req.SetBasicAuth(r.username, r.password)
	}

	for _, c := range r.cookies {
		req.AddCookie(c)
	}

	return req
}

func (r *Req) setupPool(req *http.Request) {

	r.p = &sync.Pool{
		New: func() interface{} {
			return req.Clone(context.Background())
		},
	}

	r.p.Put(req)

}

type Reqs []*Req

// NewReqs from urls
func NewReqs(us ...string) Reqs {
	tus := make(Reqs, 0, len(us))
	for _, u := range us {
		tus = append(tus, NewReq(u))
	}
	return tus
}

func (rs Reqs) SetCookies(cs []*http.Cookie) Reqs {
	for _, u := range rs {
		u.cookies = cs
	}
	return rs
}

func (rs Reqs) SetHeader(hdr http.Header) Reqs {
	for _, u := range rs {
		u.header = hdr
	}
	return rs
}

func (rs Reqs) SetUserAgent(ua string) Reqs {
	for _, u := range rs {
		u.SetUserAgent(ua)
	}
	return rs
}

func (rs Reqs) SetReferer(r string) Reqs {
	for _, u := range rs {
		u.SetReferer(r)
	}
	return rs
}

func (rs Reqs) setupPools() {
	for _, r := range rs {
		r.setupPool(r.ToRequest(http.MethodGet))
	}
}

// return a request from pool
func (rs Reqs) get() func() (int, *http.Request) {
	i := -1
	l := &sync.Mutex{}
	return func() (int, *http.Request) {
		l.Lock()
		defer l.Unlock()
		i++
		j := i % len(rs)
		return j, rs[j].p.Get().(*http.Request)
	}
}

// put a req to pool
func (rs Reqs) put(i int, req *http.Request) {
	rs[i].p.Put(req)
}

func setRange(req *http.Request, start, end int64) {
	rs := fmt.Sprintf("bytes=%d-%d", start, end-1)

	if end <= 0 {
		rs = fmt.Sprintf("bytes=%d-", start)
	}

	req.Header.Set("Range", rs)
}

func setURL(req *http.Request, u string) {
	tu, _ := url.Parse(u)
	req.URL = tu
	req.Host = tu.Host
}

func setURLName(req *http.Request, name string) {
	setURL(req, gfile.ReplaceFileName(req.URL.String(), name))
}
