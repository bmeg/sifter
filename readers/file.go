package readers

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
)

func ReadFileLines(path string) (chan []byte, error) {
	if file, err := os.Open(path); err == nil {
		return ReadLines(file)
	} else {
		return nil, err
	}
}

func ReadGzipLines(path string) (chan []byte, error) {
	if gfile, err := os.Open(path); err == nil {
		file, err := gzip.NewReader(gfile)
		if err != nil {
			return nil, err
		}
		return ReadLines(file)
	} else {
		return nil, err
	}
}

func ReadLines(r io.Reader) (chan []byte, error) {
	out := make(chan []byte, 100)
	go func() {
		reader := bufio.NewReaderSize(r, 102400)
		var isPrefix bool = true
		var err error = nil
		var line, ln []byte
		for err == nil {
			line, isPrefix, err = reader.ReadLine()
			ln = append(ln, line...)
			if !isPrefix {
				out <- ln
				ln = []byte{}
			}
		}
		close(out)
	}()
	return out, nil
}
