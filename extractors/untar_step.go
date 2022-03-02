package extractors

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"os"
	"strings"

	"github.com/bmeg/sifter/evaluate"
	"github.com/bmeg/sifter/task"
)

type UntarStep struct {
	Input string `json:"input" jsonschema_description:"Path to TAR file"`
	Strip int    `json:"strip" jsonschema_description:"Number of base directories to strip with untaring"`
}

func (us *UntarStep) Run(task *task.Task) error {
	input, err := evaluate.ExpressionString(us.Input, task.Inputs, nil)
	if err != nil {
		return err
	}
	log.Printf("Reading %s", input)
	filePath, err := task.AbsPath(input)
	if err != nil {
		return err
	}
	fhd, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fhd.Close()

	var hd io.Reader
	if strings.HasSuffix(input, ".gz") || strings.HasSuffix(input, ".tgz") {
		hd, err = gzip.NewReader(fhd)
		if err != nil {
			return err
		}
	} else {
		hd = fhd
	}

	tr := tar.NewReader(hd)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}
		//task.Printf("File: %s\n", hdr.Name)
		outPath, err := task.AbsPath(hdr.Name)
		if err != nil {
			return err
		}
		if hdr.Typeflag == tar.TypeDir {
			os.MkdirAll(outPath, 0700)
		} else if hdr.Typeflag == tar.TypeReg {
			out, err := os.Create(outPath)
			if _, err := io.Copy(out, tr); err != nil {
				//task.Printf("Failed: %s", err)
				log.Printf("Failed: %s", err)
			}
			if err != nil {
				return err
			}
		}
	}

	return nil
}
