package m3u8

import (
	"log"
	"net/http"
	"net/url"

	"github.com/grafov/m3u8"
	"github.com/nilsocket/avp"
	"github.com/nilsocket/gaja/gfile"
)

// Segments returns multiple media playlists files, (say, audio and video)
// len([][]string) = no.of selected media playlists
func Segments(c *http.Client, req *http.Request) [][]string {
	pl, lt := decode(c, req)

	switch lt {
	case m3u8.MASTER:
		return master(c, req, pl.(*m3u8.MasterPlaylist))
	case m3u8.MEDIA:
		return [][]string{media(pl.(*m3u8.MediaPlaylist))}
	}

	return nil
}

func master(c *http.Client, req *http.Request, p *m3u8.MasterPlaylist) [][]string {
	fs := masterToAVPFormats(p)
	a := avp.New(fs)
	sfs := a.Map()[avp.Best]

	ms := make([][]string, 0)

	for _, sf := range sfs {
		setURLName(req, p.Variants[sf.ID].URI)
		res := Segments(c, req)
		ms = append(ms, res[0])
	}

	return ms
}

func media(mpl *m3u8.MediaPlaylist) []string {
	segments := make([]string, 0)

	for _, s := range mpl.Segments {
		if s != nil {
			segments = append(segments, s.URI)
		}
	}

	return segments
}

func decode(c *http.Client, req *http.Request) (m3u8.Playlist, m3u8.ListType) {
	resp, err := c.Do(req)
	if err != nil {
		log.Println(err)
	}

	pl, lt, err := m3u8.DecodeFrom(resp.Body, false)
	if err != nil {
		log.Println(err)
		return nil, m3u8.MASTER
	}

	resp.Body.Close()

	return pl, lt
}

func masterToAVPFormats(p *m3u8.MasterPlaylist) avp.Formats {
	var fs avp.Formats

	for i, v := range p.Variants {

		res, _ := avp.ResolutionToInt(v.Resolution)
		fs = append(fs, &avp.Format{
			ID:           i,
			VideoBitrate: int(v.AverageBandwidth) / 1000,
			Resolution:   res,
		})
	}
	return fs
}

// copied from req.go
func setURL(req *http.Request, u string) {
	tu, _ := url.Parse(u)
	req.URL = tu
	req.Host = tu.Host
}

// copied from req.go
func setURLName(req *http.Request, name string) {
	nu := gfile.ReplaceFileName(req.URL.String(), name)
	setURL(req, nu)
}
