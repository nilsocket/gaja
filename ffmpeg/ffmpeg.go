// package ffmpeg implements helper functions wrapping around ffmpeg command.
package ffmpeg

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nilsocket/gaja/gfile"
)

var RemoveTemp = false

// M3U8SegmentsToMKV converts given tsFiles into mkv file
// [0], a_1.ts, a_2.ts, ... -> a.ts
// [1], v_1.ts, v_2.ts, ... -> v.ts
// ...
// a.ts , v.ts -> av.mkv
func M3U8SegmentsToMKV(fp string, tsFiles [][]string) {
	dir := filepath.Dir(fp)

	tns := []string{}

	for i, tsfs := range tsFiles {
		tns = append(tns, gfile.TempName(dir, "*.ts"))
		MergeTS(tns[i], tsfs)

		if RemoveTemp {
			gfile.Remove(tsfs...)
		}
	}

	err := ToMKV(fp, tns)
	if err != nil {
		log.Println("ffmpeg.ToMKV: ", err, fp, tns)
	} else {
		if RemoveTemp {
			gfile.Remove(tns...)
		}
	}
}

// ToMKV merges multiple `tsFiles` to mkv file
// ex: video ts files, audio ts files, ...
// a.ts, v.ts -> av.mkv
func ToMKV(fp string, tsFiles []string) error {
	fp = gfile.ReplaceExt(fp, "mkv")

	var args []string
	for _, tsFile := range tsFiles {
		args = append(args, "-i", tsFile)
	}

	args = append(args, "-c", "copy", fp)

	errBuf := &bytes.Buffer{}
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = errBuf
	err := cmd.Run()

	if errBuf.Len() != 0 {
		return fmt.Errorf("%s", errBuf.String())
	}

	return err
}

// MergeTS merges multiple tsFiles into one tsFile
// 1.ts, 2.ts, 3.ts, ... -> output.ts
//
// absolute paths only
func MergeTS(fp string, tsFiles []string) {
	fp = gfile.ReplaceExt(fp, "ts")
	dir := filepath.Dir(fp)

	inpf := generateInputFile(dir, tsFiles)
	if RemoveTemp {
		defer os.Remove(inpf)
	}

	errBuf := &bytes.Buffer{}
	cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", inpf, "-c", "copy", fp)
	cmd.Stderr = errBuf
	err := cmd.Run()

	d, _ := ioutil.ReadFile(inpf)
	if errBuf.Len() != 0 {
		log.Println("ffmpeg.MergeTS: ", errBuf.String(), fp, string(d))
	}

	if err != nil {
		d, _ := ioutil.ReadFile(inpf)
		log.Println("ffmpeg.MergeTS: ", err, fp, string(d))
	}
}

func generateInputFile(dir string, tsFiles []string) string {

	fp := gfile.TempName(dir, "input-*.txt")
	f, err := os.Create(fp)
	if err != nil {
		log.Println("ffmpeg.inputFile, os.Create: ", err)
	}

	for _, tsFile := range tsFiles {
		f.WriteString(
			fmt.Sprintf("file '%s'\n", tsFile),
		)
	}

	f.Close()

	return fp
}
