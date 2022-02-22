package gaja

import (
	"log"
	"math"
	"net/http"
	"os"
	"sync"

	"github.com/nilsocket/gaja/bar"
	"github.com/nilsocket/gaja/ghttp"
)

func init() {
	log.SetOutput(logf)
}

var logfName = "log.gaja"
var logf, _ = os.Create(logfName)

type Gaja struct {
	Client  *http.Client
	BaseDir string
	ConcDls int
	// Status  chan Status
	fic chan *ghttp.File
	wg  *sync.WaitGroup
}

var DefGaja = setDefaultOpts(&Gaja{})

// WithOpts ,returns *Gaja object with given opts
// unset fields are set internally
func WithOpts(opts *Gaja) *Gaja {
	if opts == nil {
		opts = &Gaja{}
	}

	return setDefaultOpts(opts)
}

// Download given set of files
// Download is blocking call, if more than `ConcDls` are in progress.
func Download(fs ...*ghttp.File) {
	DefGaja.Download(fs...)
}

// Download given set of files
// Download is blocking call, if more than `ConcDls` are in progress.
func (g *Gaja) Download(fs ...*ghttp.File) {
	for _, f := range fs {
		g.fic <- f
	}
}

// GroupCB ,calls provided `fns` once all `fs` are downloaded.
// Can be used for merging multiple files,
// or for any operation on multiple set of files.
func GroupCB(fs ghttp.Files, fns ...func(fs ghttp.Files)) ghttp.Files {
	return DefGaja.GroupCB(fs, fns...)
}

// GroupCB ,calls provided `fns` once all `fs` are downloaded.
// Can be used for merging multiple files,
// or for any operation on multiple set of files.
func (g *Gaja) GroupCB(fs ghttp.Files, fns ...func(fs ghttp.Files)) ghttp.Files {
	rcv := make(chan struct{})

	ph := func(f *ghttp.File) {
		rcv <- struct{}{}
	}

	fs.PostHooks(ph)

	// once a file is downloaded,
	// we get informed through `rcv`,
	// once all given set of files are downloaded,
	// we call provided callback functions
	n := len(fs)

	g.wg.Add(1)

	go func() {
		defer g.wg.Done()
		for range rcv {
			n--
			if n == 0 {
				close(rcv)
				break
			}
		}

		for _, fn := range fns {
			if fn != nil {
				fn(fs)
			}
		}
	}()

	return fs
}

// Close this downloader
// waits till all downloads are completed
func Close() {
	DefGaja.Close()
}

// Close this downloader
// waits till all downloads are completed
func (g *Gaja) Close() {
	close(g.fic)
	g.wg.Wait()
	bar.Wait()

	if fi, _ := logf.Stat(); fi.Size() == 0 {
		logf.Close()
		os.Remove(logfName)
	}
}

func (g *Gaja) workers() {
	g.wg.Add(g.ConcDls)

	for i := 0; i < g.ConcDls; i++ {
		go func() {
			for f := range g.fic {
				g.download(f)
			}
			g.wg.Done()
		}()
	}
}

var concDls = 4
var baseDir, _ = os.Getwd()

func setDefaultOpts(opts *Gaja) *Gaja {

	if opts.Client == nil {
		opts.Client = defClient()
	}
	if opts.ConcDls == 0 {
		opts.ConcDls = concDls
	}
	if opts.BaseDir == "" {
		opts.BaseDir = baseDir
	}

	opts.fic = make(chan *ghttp.File)
	opts.wg = new(sync.WaitGroup)

	opts.workers()

	return opts
}

func defClient() *http.Client {

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxConnsPerHost = 0
	t.MaxIdleConns = 0
	t.MaxIdleConnsPerHost = math.MaxInt32 // just to be on the safer side

	return &http.Client{
		Transport: t,
	}
}
