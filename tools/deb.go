package scans

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slices"
)

var DPKG_INFO = "/var/lib/dpkg/info"

func errhandle(e error) bool {
	if e != nil {
		fmt.Println("ERR", e)
		// os.Exit(0)
		return true
	}
	return false
}

func Deb() {
	fmt.Println("Find tracked...")
	folder, err := os.Open(DPKG_INFO)
	if errhandle(err) {
		return
	}
	files, err := folder.Readdir(0)
	if errhandle(err) {
		return
	}

	md5sum_files := make([]os.FileInfo, 0, len(files)/2)

	for _, v := range files {
		// fmt.Println(v.Name(), v.IsDir())
		if !v.IsDir() && strings.HasSuffix(v.Name(), ".md5sums") {
			md5sum_files = append(md5sum_files, v)
		}
	}

	tracked_files := make([]string, 0, len(md5sum_files)*5)

	for _, file := range md5sum_files {
		cntnt, err := os.ReadFile(DPKG_INFO + "/" + file.Name())
		if errhandle(err) {
			return
		}
		// fmt.Println(file.Name() + ":")
		lines := strings.Split(strings.TrimSpace(string(cntnt)), "\n")
		for _, line := range lines {
			fpath := "/" + strings.TrimSpace(strings.Join(strings.Split(line, " ")[1:], " "))
			abs_path, err := filepath.EvalSymlinks(string(fpath))
			if !errhandle(err) {
				tracked_files = append(tracked_files, abs_path)
			}

		}
	}
	fmt.Println("Find elfs...")
	elfs := make([]string, 0, len(md5sum_files))

	filepath.WalkDir("/", func(path string, d fs.DirEntry, err error) error {
		if errhandle(err) {
			return nil
		}

		if d.IsDir() && path == "/mnt" {
			return filepath.SkipDir
		}

		info, err := os.Lstat(path)
		if errhandle(err) {
			return nil
		}
		mode := info.Mode()
		nonRegFile := fs.ModeDir | fs.ModeSymlink | fs.ModeDevice | fs.ModeNamedPipe | fs.ModeSocket | fs.ModeCharDevice

		if mode&nonRegFile == 0 && mode&0111 != 0 {
			absPath, err := filepath.EvalSymlinks(path)
			if !errhandle(err) && absPath != path {
				return nil
			}
			// filecnt, err := os.ReadFile(path)
			opened_file, err := os.Open(path)
			if errhandle(err) {
				return nil
			}
			defer opened_file.Close()

			buf := make([]byte, 4)
			_, err = opened_file.Read(buf)
			if errhandle(err) {
				return nil
			}
			if bytes.Equal(buf, []byte{127, 69, 76, 70}) {
				// fmt.Println(path)
				elfs = append(elfs, path)
			}
		}
		return nil
	})

	// non_trk := make([]string, 0, 10)
	for _, elf := range elfs {
		if !slices.Contains(tracked_files, elf) {
			// append(non_trk, elf)
			fmt.Println(elf)
		}
	}

}
