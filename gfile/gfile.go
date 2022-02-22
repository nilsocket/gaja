//package gfile provides file and path helper functions
package gfile

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/nilsocket/svach"
)

// OpenFiles opens multiple files
// returns the first error
func OpenFiles(fs ...string) []*os.File {
	res := make([]*os.File, len(fs))
	var err error

	for i, fp := range fs {
		res[i], err = os.OpenFile(fp, os.O_CREATE|os.O_RDWR, os.ModePerm)
		if err != nil {
			return res
		}
	}

	return res
}

// Remove multiple files
func Remove(fs ...string) {
	for _, f := range fs {
		os.Remove(f)
	}
}

// TempName returns temporary name
func TempName(dir, pattern string) string {
	f, _ := os.CreateTemp(dir, "*")
	n := f.Name()
	f.Close()
	os.Remove(n)
	return n
}

// Exists returns true, if file exists
func Exists(p string) bool {
	_, err := os.Stat(p)
	return !os.IsNotExist(err)
}

// ReplaceExt with ext,
// append ext, if name doesn't have any ext
func ReplaceExt(name string, ext string) string {
	pe := filepath.Ext(name)
	if pe != "" {
		return name[:len(name)-len(pe)+1] + ext
	}
	return name + "." + ext
}

// ReplaceFileName with newname
func ReplaceFileName(oldpath, newname string) string {
	return dir(oldpath) + newname
}

func dir(path string) string {
	return path[:strings.LastIndex(path, "/")+1]
}

// IsM3U8 returns true if u ends with .m3u8
func IsM3U8(u string) bool {
	return path.Ext(u) == ".m3u8"
}

// NameFromURL returns file path without
func NameFromURL(u string) string {
	tu, _ := url.Parse(u)
	return svach.Clean(path.Base(tu.Path))
}
