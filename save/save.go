package save

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/nilsocket/gaja/gfile"
)

// https://chromedevtools.github.io/devtools-protocol/

// Future use:
// https://github.com/chromedp/chromedp/issues/659#issuecomment-674649979

var browser context.Context
var BrowserCanc context.CancelFunc

// https://github.com/chromedp/chromedp/issues/532
// https://github.com/chromedp/chromedp/blob/864094d66c22fb1ff3097a674509a165b7e74a6d/example_test.go#L200
func init() {
	browser, BrowserCanc = chromedp.NewContext(context.Background())
	chromedp.Run(browser)
}

type Type int

const (
	None Type = iota
	PDF
	IMG
	MHTML
)

func As(typ Type, req *http.Request, fp string) {
	WithActions(typ, req, fp, []chromedp.Action{}, []chromedp.Action{})
}

func WithActions(typ Type, req *http.Request, fp string, preActions, postActions []chromedp.Action) {
	if gfile.Exists(fp) {
		return
	}

	tf := tempFile(fp)

	f, _ := os.Create(tf)
	as := actions(typ, req, f, preActions, postActions)

	tab, closeTab := chromedp.NewContext(browser)

	if err := chromedp.Run(tab, as...); err != nil {
		log.Println("saveAs, err: ", err)
	}
	closeTab()

	f.Close()
	os.Rename(tf, fp)
}

func (t Type) Ext() string {
	switch t {
	case PDF:
		return "pdf"
	case MHTML:
		return "mhtml"
	case IMG:
		return "png"
	default:
		return ""
	}
}

func (t Type) action() func(*os.File) chromedp.Action {
	switch t {
	case PDF:
		return pdfAction
	case MHTML:
		return mhtmlAction
	case IMG:
		return imgAction
	default:
		return nil
	}
}

// https://github.com/chromedp/examples/blob/master/screenshot/main.go
func imgAction(f *os.File) chromedp.Action {

	return chromedp.ActionFunc(func(ctx context.Context) error {
		var buf []byte

		err := chromedp.FullScreenshot(&buf, 100).Do(ctx)
		if err != nil {
			return err
		}

		err = chromedp.ResetViewport().Do(ctx)
		if err != nil {
			return err
		}

		f.Write(buf)

		return nil
	})
}

func mhtmlAction(f *os.File) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		d, err := page.CaptureSnapshot().Do(ctx)
		if err != nil {
			return err
		}

		f.WriteString(d)

		return nil
	})
}

func pdfAction(f *os.File) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		buf, _, err := page.PrintToPDF().WithPreferCSSPageSize(true).WithMarginTop(0.4).WithMarginBottom(0.4).Do(ctx)
		if err != nil {
			return err
		}
		f.Write(buf)

		return nil
	})
}

func setHeader(h http.Header) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		th := make(network.Headers)
		for k, v := range h {
			if len(v) > 1 {
				th[k] = v
			} else {
				th[k] = h.Get(k)
			}
		}
		return network.SetExtraHTTPHeaders(th).Do(ctx)
	})
}

func actions(typ Type, req *http.Request, f *os.File, preActions, postActions []chromedp.Action) []chromedp.Action {
	defActions := []chromedp.Action{
		setHeader(req.Header), // cookies are already set in header
		chromedp.Navigate(req.URL.String()),
		chromedp.WaitReady("body"),
	}

	l := len(defActions) + len(preActions) + 1 + len(postActions)

	actions := make([]chromedp.Action, 0, l)
	actions = append(actions, defActions...)
	actions = append(actions, preActions...)
	actions = append(actions, typ.action()(f))
	actions = append(actions, postActions...)

	return actions
}

var tempExt = "-temp"

func tempFile(fp string) string {
	return fp + tempExt
}
