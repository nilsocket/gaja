package ghttp

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/nilsocket/gaja/bar"
	"github.com/nilsocket/gaja/ffmpeg"
	"github.com/nilsocket/gaja/gfile"
	"github.com/nilsocket/gaja/m3u8"
)

var m3u8DirExt = "-temp"

func m3u8Dir(fp string) string {
	return fp + m3u8DirExt
}

func (f *File) M3U8DL(c *http.Client) {
	fp := filepath.Join(f.Dir, gfile.ReplaceExt(f.Name, "mkv"))

	if gfile.Exists(fp) {
		return
	}

	f.setup()

	// multiple file segments
	mfs := getSegmentIDs(c, f)
	mfsLen := int64(mfsLen(mfs))

	tDir := m3u8Dir(fp)
	os.MkdirAll(tDir, os.ModePerm)
	log.Println(fp)

	b := bar.M3U8(f.Name, mfsLen)
	wg := new(sync.WaitGroup)
	segChan := make(chan string)

	// init done
	runHooks(f, f.PreHooks)
	defer runHooks(f, f.PostHooks)

	wg.Add(1)
	go sendSegmentIDs(mfs, segChan, wg)

	wg.Add(f.ConcReqs)
	for i := 0; i < f.ConcReqs; i++ {
		go func() {
			defer wg.Done()

			for segID := range segChan {
				fp := filepath.Join(tDir, segID)

				if gfile.Exists(fp) {
					if b != nil {
						b.SetRefill(1)
					}
					continue
				}

				j, req := f.getReq()
				dlSegment(c, req, fp, segID)
				f.putReq(j, req)

				if b != nil {
					b.Increment()
				}
			}
		}()
	}

	wg.Wait()

	// modify segments name
	for _, segs := range mfs {
		for j, segID := range segs {
			segs[j] = filepath.Join(tDir, segID)
		}
	}

	ffmpeg.M3U8SegmentsToMKV(fp, mfs)

	os.RemoveAll(tDir)
}

// dlSegment
func dlSegment(c *http.Client, req *http.Request, fp, segID string) {

	if gfile.Exists(fp) {
		return
	}

	sfp := sequentialFile(fp)
	dlf := gfile.OpenFiles(sfp)[0]

	setURLName(req, segID) // change url's file name to segID

	if gfile.Exists(sfp) {
		asHead(req)
		resumable, _ := ras(c, req)

		if resumable {
			fi, _ := dlf.Stat()
			cs := fi.Size()
			setRange(req, cs, 0)
			dlf.Seek(cs, io.SeekStart)
		}
	}

	asGet(req)
	seqDL(c, req, dlf, nil)
	dlf.Close()
	os.Rename(sfp, fp)

}

func asHead(req *http.Request) {
	req.Method = http.MethodHead
}

func asGet(req *http.Request) {
	req.Method = http.MethodGet
}

func getSegmentIDs(c *http.Client, f *File) [][]string {
	i, req := f.getReq()
	mfs := m3u8.Segments(c, req)
	f.putReq(i, req)
	return mfs
}

func sendSegmentIDs(mfs [][]string, segChan chan string, wg *sync.WaitGroup) {
	for _, fs := range mfs {
		for _, seg := range fs {
			segChan <- seg
		}
	}

	close(segChan)
	wg.Done()
}

func mfsLen(mfs [][]string) int {
	l := 0
	for _, fs := range mfs {
		l += len(fs)
	}
	return l
}
