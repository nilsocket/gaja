package gaja

import (
	"os"
	"path/filepath"

	"github.com/nilsocket/gaja/gfile"
	"github.com/nilsocket/gaja/ghttp"
)

func (g *Gaja) download(f *ghttp.File) {

	setF(g, f)

	if f.SaveOpts != nil {
		f.Save()
		return
	}

	if gfile.IsM3U8(f.Reqs[0].URL) {
		f.M3U8DL(g.Client)
		return
	}

	f.Download(g.Client)
}

func setF(g *Gaja, f *ghttp.File) {
	if f.Name == "" {
		f.Name = gfile.NameFromURL(f.Reqs[0].URL)
	}

	if f.Dir != "" {
		f.Dir = filepath.Join(g.BaseDir, f.Dir)
	} else {
		f.Dir = g.BaseDir
	}

	os.MkdirAll(f.Dir, os.ModePerm)

	if f.ConcReqs == 0 {
		f.ConcReqs = 3
	}
}
