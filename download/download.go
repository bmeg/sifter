package download

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-getter"
	"github.com/jlaffaye/ftp"
)

func ToFile(src string, dest string) (string, error) {

	var err error
	dest, err = filepath.Abs(dest)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(src, "ftp:") {
		u, err := url.Parse(src)
		if err != nil {
			return "", err
		}
		c, err := ftp.Dial(u.Host+":21", ftp.DialWithTimeout(5*time.Second))
		if err != nil {
			return "", err
		}
		err = c.Login("anonymous", "anonymous")
		if err != nil {
			return "", err
		}
		r, err := c.Retr(u.Path)
		if err != nil {
			return "", err
		}
		defer r.Close()

		log.Printf("Saving to %s", dest)
		f, err := os.Create(dest)
		if err != nil {
			return "", err
		}
		defer f.Close()

		downloadSize, _ := io.Copy(f, r)
		log.Printf("Downloaded %d bytes", downloadSize)
		return dest, nil
	}

	if strings.HasPrefix(src, "s3:") {
		s3Key := os.Getenv("AWS_ACCESS_KEY_ID")
		s3Secret := os.Getenv("AWS_SECRET_ACCESS_KEY")
		s3Endpoint := os.Getenv("AWS_ENDPOINT")
		if s3Endpoint != "" {
			u, err := url.Parse(src)
			if err != nil {
				return "", err
			}
			//"s3::http://127.0.0.1:9000/test-bucket/hello.txt?aws_access_key_id=KEYID&aws_access_key_secret=SECRETKEY&region=us-east-2"
			src = fmt.Sprintf("s3::%s/%s%s", s3Endpoint, u.Host, u.Path)
		}
		if s3Key != "" && s3Secret != "" {
			src = src + fmt.Sprintf("?aws_access_key_id=%s&aws_access_key_secret=%s", s3Key, s3Secret)
		}
		src = src + "&archive=false"
	} else {
		src = src + "?archive=false"
	}

	return dest, getter.GetFile(dest, src)
}
