package ghttp

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/nilsocket/gaja/bar"
	"github.com/nilsocket/gaja/gfile"
	"github.com/nilsocket/gaja/gio"
	"github.com/nilsocket/gaja/scheduler"
	"github.com/vbauerster/mpb/v6"
)

var progressExt = ".pgaja"
var sequentialExt = ".sgaja"

var pool = &sync.Pool{
	New: func() interface{} {
		tbuf := make([]byte, 5*scheduler.MiB)
		return &tbuf
	},
}

func (f *File) Download(c *http.Client) {
	fp := filepath.Join(f.Dir, f.Name)
	pfp := progressFile(fp)

	if completed(fp) {
		return
	}

	resumable, size := resumableAndSize(c, f)

	if !resumable {
		f.SequentialDL(c)
		return
	}

	b := bar.New(f.Name, size)
	s := scheduler.NewStreaming(b, pfp, 0, size)

	f.setup()
	dlf := gfile.OpenFiles(fp)[0]
	wg := new(sync.WaitGroup)

	// init done
	runHooks(f, f.PreHooks)
	defer runHooks(f, f.PostHooks)

	wg.Add(f.ConcReqs)

	for i := 0; i < f.ConcReqs; i++ {
		go func() {
			defer wg.Done()

			for fr := s.Range(); fr != nil; fr = s.Range() {

				j, req := f.getReq()
				setRange(req, fr.Start, fr.End)
				resp, err := c.Do(req)
				if err != nil {
					log.Printf("ghttp.f.Download, c.Do(%s), %s, err: %s\n", req.URL, fr.String(), err)

					f.putReq(j, req)
					s.Failed(fr)
					time.Sleep(time.Millisecond * 100) // maybe too many connections
					continue
				}

				f.copyAt(s, fr, dlf, resp.Body)
				resp.Body.Close()
				f.putReq(j, req)
			}
		}()
	}

	wg.Wait()
	dlf.Close()
	s.Close()
	os.Remove(pfp)
}

func (f *File) setup() {
	f.Reqs.setupPools()
	f.getReq = f.Reqs.get()
	f.putReq = f.Reqs.put
}

func (f *File) copyAt(s *scheduler.Streaming, fr *scheduler.Range, dst *os.File, src io.ReadCloser) {
	buf := pool.Get().(*[]byte)

	sw := gio.NewSectionWriter(dst, fr.Start, fr.End,
		func(from, to int64) { s.Done(&scheduler.Range{Start: from, End: to}) },
		func(from, to int64) { s.Failed(&scheduler.Range{Start: from, End: to}) },
	)

	io.CopyBuffer(sw, src, *buf)

	pool.Put(buf)
}

func (f *File) SequentialDL(c *http.Client) {
	fp := filepath.Join(f.Dir, f.Name)

	if gfile.Exists(fp) {
		return
	}

	sfp := sequentialFile(fp)
	dlf := gfile.OpenFiles(sfp)[0]
	req := f.Reqs[0].ToRequest(http.MethodGet)
	resumable, size := resumableAndSize(c, f)
	b := bar.New(path.Base(fp), size)

	if gfile.Exists(sfp) {
		if resumable {
			fi, _ := dlf.Stat()
			cs := fi.Size()
			setRange(req, cs, 0)
			b.SetRefill(cs)
			dlf.Seek(cs, io.SeekStart)
		}
	}

	// init done
	runHooks(f, f.PreHooks)
	defer runHooks(f, f.PostHooks)

	seqDL(c, req, dlf, b)
	dlf.Close()
	os.Rename(sfp, fp)
}

func seqDL(c *http.Client, r *http.Request, dlf *os.File, b *mpb.Bar) {
	for err := iSeqDL(c, r, dlf, b); err != nil; err = iSeqDL(c, r, dlf, b) {
		dlf.Seek(0, io.SeekStart)
		if b != nil {
			b.SetCurrent(0)
		}
	}
}

func iSeqDL(c *http.Client, r *http.Request, dlf *os.File, b *mpb.Bar) error {

	resp, err := c.Do(r)
	if err != nil {
		log.Printf("ghttp.iseqDl, c.Do(%s), err: %s \n", r.URL, err)
		return err
	}

	pr := resp.Body
	if b != nil {
		pr = b.ProxyReader(resp.Body)
	}

	buf := pool.Get().(*[]byte)
	_, err = io.CopyBuffer(dlf, pr, *buf)
	pool.Put(buf)
	resp.Body.Close()

	return err
}

// TODO: other methods to detect
func resumableAndSize(c *http.Client, f *File) (bool, int64) {
	return ras(c, f.Reqs[0].ToRequest(http.MethodHead))
}

func ras(c *http.Client, headReq *http.Request) (bool, int64) {
	resp, err := c.Do(headReq)
	if err != nil {
		log.Println("ghttp.resumableAndSize, c.Do(): ", err)
		return false, 0
	}

	if resp.Header.Get("Accept-Ranges") == "bytes" {
		return true, resp.ContentLength
	}

	return false, resp.ContentLength
}

func progressFile(fp string) string {
	return fp + progressExt
}

func sequentialFile(fp string) string {
	return fp + sequentialExt
}

func completed(fp string) bool {
	return gfile.Exists(fp) && !gfile.Exists(progressFile(fp))
}

func runHooks(f *File, hs []func(*File)) {
	for _, h := range hs {
		h(f)
	}
}
