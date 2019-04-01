package manager

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"os"
	"strings"

	"github.com/bmeg/sifter/evaluate"
)

type UntarStep struct {
	Input string `json:"input"`
	Strip int    `json:strip`
}

func (us *UntarStep) Run(task *Task) error {
	input, err := evaluate.ExpressionString(us.Input, task.Inputs)
	if err != nil {
		return err
	}
	log.Printf("Reading %s", input)
	filePath, err := task.Path(input)
	if err != nil {
		return err
	}
	var hd io.Reader
	hd, err = os.Open(filePath)
	if err != nil {
		return err
	}

	if strings.HasSuffix(input, ".gz") || strings.HasSuffix(input, ".tgz") {
		hd, err = gzip.NewReader(hd)
		if err != nil {
			return err
		}
	}
	//defer hd.Close()

	tr := tar.NewReader(hd)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}
		task.Printf("File: %s\n", hdr.Name)
		outPath, err := task.Path(hdr.Name)
		if err != nil {
			return err
		}
		if hdr.Typeflag == tar.TypeDir {
			os.MkdirAll(outPath, 0700)
		} else if hdr.Typeflag == tar.TypeReg {
			out, err := os.Create(outPath)
			if _, err := io.Copy(out, tr); err != nil {
				task.Printf("Failed: %s", err)
			}
			if err != nil {
				return err
			}
		}
	}

	return nil
}
