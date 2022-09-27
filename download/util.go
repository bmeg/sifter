package download

import (
	"strings"
)

func IsURL(s string) bool {
	if strings.HasPrefix(s, "http://") {
		return true
	}
	if strings.HasPrefix(s, "https://") {
		return true
	}
	if strings.HasPrefix(s, "s3://") {
		return true
	}
	if strings.HasPrefix(s, "ftp://") {
		return true
	}
	return false
}
