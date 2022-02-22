package bar

import (
	"strings"
	"time"

	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
)

var p *mpb.Progress

// maxNameSize while displaying name on bar
var maxNameSize = 30

func init() {
	p = mpb.New(mpb.WithRefreshRate(time.Millisecond * 500))
}

// var barStyle = " ██  █▒"
var barStyle2 = "[##-]#-"

func New(name string, length int64) *mpb.Bar {
	return newBar(name, false, length)
}

func M3U8(name string, length int64) *mpb.Bar {
	return newBar(name, true, length)
}

func Spinner(name string) *mpb.Bar {
	tempName := shortName(name, maxNameSize-5)

	return p.Add(1, mpb.NewSpinnerFiller(spinnerStyle, mpb.SpinnerOnLeft),
		mpb.PrependDecorators(
			decor.OnComplete(
				decor.Name(tempName, decor.WC{W: maxNameSize + 2, C: decor.DidentRight}),
				name,
			),
		),
		mpb.BarFillerClearOnComplete(),
	)
}

func Wait() {
	p.Wait()
}

func newBar(name string, m3u8 bool, length int64) *mpb.Bar {
	tempName := shortName(name, maxNameSize-5)

	dec := normalDecorators(tempName, name)
	if m3u8 {
		dec = m3u8Decorators(tempName, name)
	}

	return p.Add(length, mpb.NewBarFiller(barStyle2),
		mpb.PrependDecorators(dec...),
		mpb.AppendDecorators(
			decor.OnComplete(
				decor.Percentage(decor.WC{W: 5}), ""),
		),
		mpb.BarFillerClearOnComplete(),
	)
}

func m3u8Decorators(tempName, name string) []decor.Decorator {
	return []decor.Decorator{
		decor.OnComplete(
			decor.Name(tempName, decor.WC{W: maxNameSize + 2, C: decor.DidentRight}),
			name,
		),
		decor.OnComplete(
			decor.CountersNoUnit("%d/%d", decor.WC{W: 18, C: decor.DidentRight}),
			"",
		),
	}
}

func normalDecorators(tempName, name string) []decor.Decorator {
	return []decor.Decorator{
		decor.OnComplete(
			decor.Name(tempName, decor.WC{W: maxNameSize + 2, C: decor.DidentRight}),
			name,
		),
		decor.OnComplete(
			decor.AverageSpeed(decor.UnitKB, "%.2f", decor.WC{W: 11, C: decor.DidentRight}),
			"",
		),
		decor.OnComplete(
			decor.CountersKiloByte("%.2f/%.2f", decor.WC{W: 18, C: decor.DidentRight}),
			"",
		),
	}
}

var continuation = "..."

// shortName returns a shortened string of size max
func shortName(s string, max int) string {

	if len(s) > max {
		if i := strings.LastIndex(s, "."); i > 0 {
			// >7 , most probably not extension
			if len(s)-i > 7 {
				return s[:max-3] + continuation
			}
			return s[:max-3] + continuation + s[i:]
		}
		return s[:max-3] + continuation
	}

	return s
}
