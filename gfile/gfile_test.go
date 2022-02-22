package gfile_test

import (
	"testing"

	"github.com/nilsocket/gaja/gfile"
)

var replaceExtInp = [][]string{
	{"hello.m3u8", "mkv", "hello.mkv"},
	{"xyz.en...vtt", "mp4", "xyz.en...mp4"},
	{"xyz.mkv", "mkv", "xyz.mkv"},
	{"xyz", "mkv", "xyz.mkv"},
}

func TestReplaceExt(t *testing.T) {
	for _, inp := range replaceExtInp {
		out := gfile.ReplaceExt(inp[0], inp[1])
		if out != inp[2] {
			t.Errorf("ReplaceExt(%s, %s), Expected: %s, Got: %s", inp[0], inp[1], inp[2], out)
		}
	}
}

var urlsIO = [][]string{
	{"https://r2---sn-cnoa-itqe.googlevideo.com/videoplayback?expire=1620545806&ei=rjyXYJTmM4Gpz7sPhOW88AY&ip=117.254.149.61&id=o-AM3VsXzqXNXXuYks6ScjvQ0MMJsI74aCoUFuYbaHsAQr&itag=248&aitags=133%2C134%2C135%2C136%2C137%2C160%2C242%2C243%2C244%2C247%2C248%2C278&source=youtube&requiressl=yes&mh=Sj&mm=31%2C29&mn=sn-cnoa-itqe%2Csn-h557sns7&ms=au%2Crdu&mv=m&mvi=2&pl=21&initcwndbps=526250&vprv=1&mime=video%2Fwebm&ns=O9Gx4tu0PvjIpJ42XXbjHsQF&gir=yes&clen=538741975&dur=3989.418&lmt=1586149480146621&mt=1620523713&fvip=2&keepalive=yes&fexp=24001373%2C24007246&beids=9466588&c=WEB&txp=5511222&n=QWCtZyAIzMTeZUg7j_B&sparams=expire%2Cei%2Cip%2Cid%2Caitags%2Csource%2Crequiressl%2Cvprv%2Cmime%2Cns%2Cgir%2Cclen%2Cdur%2Clmt&sig=AOq0QJ8wRAIgAZTVBTdV5BTW1KRsXQ1t59PMueZsejhjatX0rOXQRDQCIFdH6lQor3_hodGLsFN493NbIrYd8ecyc9bO_e0pIZYu&lsparams=mh%2Cmm%2Cmn%2Cms%2Cmv%2Cmvi%2Cpl%2Cinitcwndbps&lsig=AG3C_xAwRQIhAKsu0bwwMdSebJsTGdFLggqLCqekB0IvXs9YXbd_AP7AAiAvxzWHbTvfnXqr5tZrDPBwAZ9XjblC4WU6unxeZgG0cA%3D%3D", "videoplayback"},
	{"https://someurl.com/hello/world/to/the/world", "world"},
}

func TestNameFromURL(t *testing.T) {
	for _, inp := range urlsIO {
		out := gfile.NameFromURL(inp[0])
		t.Log(out)
		if out != inp[1] {
			t.Errorf("NameFromURL(%s), Expected: %s, Got: %s", inp[0], inp[1], out)
		}
	}
}

var m3u8Inp = []string{
	"https://edx-video.net/IsraelXinfosec101-V005600/IsraelXinfosec101-V005600.m3u8",
}

var m3u8Out = []bool{
	true,
}

func TestIsM3U8(t *testing.T) {
	for i, inp := range m3u8Inp {
		out := gfile.IsM3U8(inp)
		if out != m3u8Out[i] {
			t.Errorf("IsM3U8(%s), Expected: %v, Got: %v", inp, m3u8Out[i], out)
		}
	}
}
