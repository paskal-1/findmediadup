package main

import (
	"flag"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/sahilm/fuzzy"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func parseNumberWithSuffix(s string) (float64, error) {
	var multiplier float64 = 1
	switch s[len(s)-1] {
	case 'K', 'k':
		multiplier = 1e3
		s = s[:len(s)-1]
	case 'M', 'm':
		multiplier = 1e6
		s = s[:len(s)-1]
	case 'G', 'g':
		multiplier = 1e9
		s = s[:len(s)-1]
	case 'T', 't':
		multiplier = 1e12
		s = s[:len(s)-1]
	}
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return value * multiplier, nil
}

type Files struct {
	files      []File
	limitScore int
}

type File struct {
	name string
	path string
	size uint64
	root int
}

func (t File) String() string {
	return fmt.Sprintf("(%d)%s\t%s", t.root, t.path, t.name)
}

func (t *Files) addFile(path string, size uint64, root int) {
	newFile := File{filepath.Base(path), filepath.Dir(path), size, root}
	for _, oldFile := range t.files {
		matches := fuzzy.Find(newFile.name, []string{oldFile.name})
		if len(matches) > 0 && matches[0].Score > t.limitScore {
			fmt.Printf("%d\t%s\t%s/%s\t%s\t%s\n", len(t.files), humanize.SIWithDigits(float64(matches[0].Score), 0, ""), humanize.Bytes(size), humanize.Bytes(oldFile.size), oldFile, newFile)
		}
	}
	t.files = append(t.files, newFile)
}

func main() {
	limitScore := flag.Int("limitScore", 0, "minimum score to consider match")
	limitSize := flag.String("limitSize", "0", "minimum file size to consider match, like 1G")
	flag.Parse()

	limitSizeNum, err := parseNumberWithSuffix(*limitSize)
	if err != nil {
		log.Panic("can not parse limitSize", err)
	}
	cnt := 0
	var size int64 = 0
	files := Files{limitScore: *limitScore}

	for rootNo, root := range flag.Args() {

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			path = strings.TrimPrefix(path, root)
			if err != nil {
				fmt.Println(err)
			}
			cnt++
			size += info.Size()
			if !info.IsDir() && float64(info.Size()) >= limitSizeNum {
				files.addFile(path, uint64(info.Size()), rootNo)
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error processing %s : %s", root, err)
		}
		fmt.Fprintf(os.Stderr, "root %d - visiting %d analyzing %d, scanned %s\n", rootNo, cnt, len(files.files), humanize.Bytes(uint64(size)))
	}
	fmt.Fprintf(os.Stderr, "Finished after visiting %d and analyzing %d, scanned %s\n", cnt, len(files.files), humanize.Bytes(uint64(size)))

}
