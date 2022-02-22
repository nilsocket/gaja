package ghttp

import (
	"net/http"
	"path/filepath"

	"github.com/chromedp/chromedp"
	"github.com/nilsocket/gaja/bar"
	"github.com/nilsocket/gaja/gfile"
	"github.com/nilsocket/gaja/save"
)

type SaveOpts struct {
	Type        save.Type
	PreActions  []chromedp.Action
	PostActions []chromedp.Action
}

// Save opens given file in chrome and
// lets you save in pdf, png, mhtml format
func (f *File) Save() {
	opts := f.SaveOpts
	if opts != nil && opts.Type != save.None {
		fp := filepath.Join(f.Dir, gfile.ReplaceExt(f.Name, opts.Type.Ext()))
		b := bar.Spinner(f.Name)

		runHooks(f, f.PreHooks)
		defer runHooks(f, f.PostHooks)

		save.WithActions(opts.Type, f.Reqs[0].ToRequest(http.MethodGet), fp, opts.PreActions, opts.PostActions)
		b.Increment()
	}
}
