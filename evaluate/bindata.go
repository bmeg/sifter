// Code generated by go-bindata. DO NOT EDIT.
// sources:
// exec_pb2.py
// exec_pb2_grpc.py
// sifter-exec.py

package evaluate

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  fileInfoEx
}

type fileInfoEx interface {
	os.FileInfo
	MD5Checksum() string
}

type bindataFileInfo struct {
	name        string
	size        int64
	mode        os.FileMode
	modTime     time.Time
	md5checksum string
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) MD5Checksum() string {
	return fi.md5checksum
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _bindataExecpb2Py = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x59\xff\x6b\xdb\x46\x14\xff\x5d\x7f\xc5\x2d\x61\xb3\xdd\xb9\xc2\xdf\xe2\xa6\x01\x41\xc1\x76\x47\xa0\x6d\x42\x9a\xc2\xa0\x2a\xe2\x24\x3d\xbb\x62\xe7\x3b\xa1\x3b\x15\x65\x63\xff\xfb\xb8\x2f\xb2\x4e\xb2\xad\x28\xa5\xcb\xe8\x1a\x41\x82\xf5\xee\x7d\x7f\x4f\x9f\x7b\x3a\x9d\xa2\xe7\xcf\x9e\xa3\x88\xc5\x09\xdd\x5c\xa0\x5c\xac\x9f\x9f\x4b\x8a\x73\x8a\x7e\x03\x0a\x19\x16\x10\xa3\xf0\x0e\x89\xcf\x80\xd2\x8c\x09\x16\x31\x82\xc2\x7c\xbd\x86\x0c\x45\x6c\x9b\x26\x04\x32\x17\xa1\xe5\x15\x7a\x77\x75\x8b\x56\xcb\xcb\xdb\x9f\x9c\x53\xc4\x59\x9e\x45\x70\x81\xa0\x80\xc8\x55\x62\x8e\xb3\xce\xd8\x16\x6d\x18\xdb\x10\xd0\xa4\x30\x5f\xa3\x64\x9b\xb2\x4c\xa0\x18\x78\x94\x25\xa9\x60\x19\xc2\x1c\x05\xd5\x6d\xab\xd4\x16\x38\xc7\x1b\x50\x22\xe6\x77\x2b\x7f\x06\x6b\x02\x91\x48\x18\x55\x22\xd5\x6d\xab\x14\xbf\xdb\x86\x8c\x04\x31\x16\x38\xc4\x5c\x5b\x6b\xd0\x9c\x53\xf4\xea\x95\xce\x4e\x90\x50\x0e\x99\x54\x1a\xa4\x2c\xa1\xa2\xaf\xb5\xf0\x81\xe3\x48\xa9\x20\x0e\x91\xb7\x27\xef\x2e\x61\x8d\x73\x22\xfa\x03\x47\x5e\xcb\xd5\xfb\xc5\xcd\xe5\xf5\xed\xd5\x8d\xe4\xad\x92\xe1\xbe\x4e\x08\x2c\x77\xb7\x7d\x07\x21\x8a\xb7\xe0\xf5\xaa\x34\xf7\x86\x0e\x42\x29\x8e\xfe\xc0\x1b\x49\xff\x82\x49\x8e\x05\x28\x2a\xbf\xa3\x02\x17\x5e\x4f\xf1\x4d\x35\x09\xb2\x04\x93\xe4\x4f\x88\x03\x96\x4a\x9f\xb9\xf7\x8e\x51\x68\x2c\xa5\xa1\x17\xf6\x7c\xea\xd3\xca\x8c\x5f\x8c\x27\x7e\x31\x3a\xf7\x8b\xf9\x99\xb1\xe1\x9f\xfc\xe2\x53\xbf\x18\xcd\xfc\x62\x36\x65\x31\x18\x96\xa8\x24\xce\x0d\xf1\xdc\x2f\x46\x63\x24\xff\xf5\x7d\xa1\x99\xc6\x23\xc5\x24\xb5\xcd\x73\xaa\x2a\x62\x18\x27\x3b\xc6\x93\x67\x3e\xf5\xb3\x85\xee\xb8\x1b\xe0\x39\xd1\xc2\x54\x89\x4e\x92\xb8\xa1\x3a\x53\xab\x99\x5a\x3d\x93\x6e\x66\x19\xcb\xf6\xb5\x9e\x6a\x86\x4b\x9a\xe6\x62\xcf\x63\xf9\x37\x16\xf8\xb0\xd3\x87\x23\x9b\xec\xcc\x9f\xfc\xac\xd6\xe7\x96\xaf\x5d\x75\xdf\xe3\xf5\x64\x5b\xa6\x6b\x76\x56\x40\x94\x0b\xa6\x83\x2d\xa6\x33\xb5\xf0\x42\x55\x40\x25\xca\x98\x05\xb7\xec\x04\x77\xa1\x7d\x95\x86\x5f\xd8\x54\x3b\xaf\x27\x7e\x31\x1a\x49\xd1\x5f\xab\x8a\x4a\x77\x09\x31\xfa\xd6\x95\x64\x99\x39\xa9\x70\x54\x91\xeb\x9a\xe6\x52\x68\x6e\x1a\xcf\xd1\x4d\x1e\x2c\xae\x96\xab\x46\x7f\x1f\xea\x6d\xe9\xb0\x6a\xd6\x75\x4e\x48\x60\xfa\xdd\x8e\x46\x2f\x26\x04\xd4\x5a\xd9\xbf\x92\xe0\x55\x0f\x92\x24\x45\x8c\x0a\x9c\xd0\x84\x6e\x02\x71\x97\xda\xac\x40\x62\xee\x7d\x74\x10\x42\x8d\xc7\x0d\x48\x5c\xf7\x49\x5e\xda\x87\x48\x99\x3e\xea\x95\x6b\xd6\x13\x1a\x43\xe1\x8d\x86\xa5\x6c\xbe\x0d\x21\xf3\xc6\x43\xa4\x5c\x78\x39\x44\x51\x9a\x06\xe5\x6f\x82\x43\x20\xde\xb8\x64\xfe\x8c\x79\x10\x6b\x64\x08\xa4\x6e\xf0\x5e\x63\xc2\x61\x88\xea\xc4\xf0\xe4\xc4\x8d\x41\xda\xeb\xf7\x14\x7e\xf7\x06\xa5\x06\x83\x8a\x56\xbc\x08\x68\xbe\xb5\xef\x8f\x65\x45\x5e\x09\x0f\xa0\x10\x40\x79\xc2\x68\x69\x7b\x47\x08\x78\xc4\x1a\xfc\xc7\xf0\xa4\x59\x0d\xe3\x5f\xe7\x5c\xaf\x0d\x2a\xb4\xe5\xdb\xe2\xd1\x39\x1f\x37\x72\x3e\x79\xca\x39\xfa\x24\xff\xed\xb4\xe9\x9e\x57\x34\x0a\x5c\x40\xac\xbc\xe1\xde\x47\xcd\x56\x3a\x6d\xb1\xb5\x6c\x18\xa5\xdf\x31\x0e\x49\x99\xb4\xc3\xbb\x4e\x15\x4d\x86\xe9\x66\x67\x8f\x51\x60\xeb\xc3\xb6\xb8\xc0\x99\xf0\x26\xb3\x06\x19\x68\xec\xcd\x27\x43\x05\x28\xc1\xe2\xea\xed\xf5\xe5\x9b\xd5\xcd\xea\xfd\x87\x37\xb7\x9d\x70\xc5\x82\xbc\x16\x80\xd9\xe3\x7a\x74\xa4\x49\xe2\xa3\x7d\x6f\x39\xe7\x2a\xb6\x76\xb8\x19\x4f\xad\xde\x9f\x7e\x6d\xef\x8f\xbe\x87\x4e\x7f\x40\x7e\x41\x6e\xb2\x9d\x52\x5c\x72\x3e\x01\xcc\xff\x12\x60\xe6\x87\x00\x66\x3c\x9a\x1b\x84\xb9\x7c\x77\xfd\xa1\x0b\xb2\xa8\x91\xe8\x28\xa2\x58\xab\x8f\x8e\x24\xf2\x7d\xe3\x48\xa3\x2b\xb7\x5c\xc3\xf0\x34\xb4\x7c\x03\x58\x69\x19\x10\x75\xb2\x6b\x13\xe2\x31\x30\xf9\x81\x20\xfb\xbb\xc6\x8e\xf1\xe8\xfc\x10\x78\xcc\xa6\x06\x3c\x3a\xcf\x25\xf7\x0c\x24\xff\xe9\x24\xd2\x82\x1f\x66\x87\x7c\x02\x90\xc7\x99\x4b\x9e\x06\x92\x1f\x02\x54\x66\x67\x87\x40\xe5\x5c\xbf\xf3\x54\x91\xba\x76\x05\x78\x10\xde\xa9\x76\xf9\xa8\xcf\x4e\x3e\x49\xd0\x59\x5c\x2d\x57\x9d\x04\xec\xd7\x1d\x23\x69\xbd\x57\x75\x50\xa1\xe7\x1b\x25\xaa\x06\xa6\x0e\x22\xb6\x39\x63\xc7\x9c\x95\xba\x37\xb0\x49\xb8\x80\xac\x71\xf2\x69\x15\xd9\x71\x64\x94\x52\xb4\x3a\xcf\x75\x77\x47\xd7\xd7\xe6\xc8\xfa\xad\xb6\x7c\x7b\x97\x42\xdf\x1c\x29\xa1\x7e\x79\x68\xec\x9a\xd5\xe1\x60\x88\xfe\x72\x10\xea\x55\xea\x7b\xe8\x42\x27\x4f\xd6\xa1\x17\x04\x5b\x16\xe7\x04\x82\x40\xd2\xd5\xb1\x6b\x90\x86\x93\x9e\x83\x50\xcb\x09\x70\x44\x30\xe7\xba\x8f\x2f\x6a\x67\x16\x03\x07\xa1\xbf\x07\x7b\xb1\x1a\x6f\xfa\x9a\xc5\xa9\xd5\xe4\x81\x71\xd6\x5e\x5e\x1f\x12\xb0\x55\xf3\x6f\x1f\xb9\xe5\xd5\xbd\x29\xa8\xf1\x3a\xaa\xb9\x1e\x94\x03\x33\x6e\x77\x8f\x5d\x35\xed\xb7\x8e\x59\x79\x71\x4f\xac\x86\xc7\xf9\x8a\x42\x3f\xbc\xc2\xff\x4e\x69\x3b\xd5\x74\x57\x4c\xc7\x09\x56\xbf\xaf\x16\x1f\xf6\xbf\x73\xbc\x87\xec\x4b\x12\x1d\xfc\xd4\xb1\x32\x27\xde\x47\x07\xa4\x3a\xc3\xfe\x44\x64\x8d\x26\xdd\x3e\x7f\x18\x28\x3e\x3f\xf4\x72\x38\x79\x39\x95\xd4\x2d\x88\xcf\xcc\x8c\x53\x76\x1c\x6f\x15\xbd\xb9\xbf\xd7\xce\x9f\x7a\x7a\x57\x6b\x8b\xc4\xad\xb3\xd6\x66\x2b\x6b\x5f\xe5\x3a\x69\xd6\x56\x99\xc8\x8e\xd2\xfb\xed\x0e\xc3\x10\x62\xb9\xb0\xc9\x8d\x27\xbd\x35\x2b\x6a\x47\xed\x1c\x20\x26\xa4\x53\x74\x15\x5f\x6d\x8c\xe9\x1c\xda\xee\x89\x6d\xc4\xd6\x3d\xa8\x4f\xfb\xdd\xba\xdf\x82\xbb\x66\xad\x6f\xbd\xc6\x37\x6b\x3b\xdb\x75\xa0\xda\xd0\x4a\x29\xa7\xed\x23\xa1\x79\x02\xd5\xe3\x34\x70\xfe\x09\x00\x00\xff\xff\x11\x1b\x02\x1e\x93\x1d\x00\x00")

func bindataExecpb2PyBytes() ([]byte, error) {
	return bindataRead(
		_bindataExecpb2Py,
		"exec_pb2.py",
	)
}

func bindataExecpb2Py() (*asset, error) {
	bytes, err := bindataExecpb2PyBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name:        "exec_pb2.py",
		size:        7571,
		md5checksum: "",
		mode:        os.FileMode(420),
		modTime:     time.Unix(1587596493, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataExecpb2grpcPy = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x56\x4d\x6f\xe3\x36\x10\xbd\xeb\x57\xcc\x7a\x0f\x96\x01\x41\x0b\xf4\x18\xc0\x40\x03\xc7\x2d\x0c\x6c\xbc\x46\xe2\x02\xbd\x11\x0c\x39\xb6\x59\x50\xa4\x3a\x1c\xa5\x49\x8b\xfe\xf7\x82\xfa\x70\x6c\xc9\xc9\xc6\x4d\xda\x53\x74\x10\x44\x71\x3e\x34\x8f\x6f\x9e\xe6\x33\xfc\x8c\x0e\x49\x32\x6a\xb8\x7b\x04\xde\x21\x6c\x6f\x56\x33\x58\x3d\xf2\xce\x3b\x28\xc9\xb3\x57\xde\x82\xf2\x45\x69\x2c\x12\x94\xb6\xda\x1a\x97\xc3\xd5\x37\x58\x7e\x5b\xc3\xfc\x6a\xb1\xfe\x94\x98\xa2\xf4\xc4\xb0\xa5\x52\x25\xdd\x02\x1f\x50\x89\xf2\xee\x07\x90\xa1\x79\x8e\x8b\x24\x49\x94\x95\x21\xc0\xfc\x01\x55\xc5\x9e\x6e\xb9\xba\x4b\xfd\xdd\x6f\xa8\x78\x72\x91\x00\x00\x8c\x46\xa3\x6b\x13\x82\x71\x5b\x90\x21\x78\x65\xea\x6f\xd3\x5e\x55\x05\x3a\x96\x6c\xbc\x8b\x5f\x13\x17\x60\x1c\xe4\xf5\x27\xc2\xc6\x58\x1c\x8d\x46\x49\x1d\x42\xe3\x06\x84\x30\xce\xb0\x10\x69\x40\xbb\xc9\x40\xed\xa4\x73\x68\xdb\x1c\x6d\x9e\x99\x77\x81\xa9\x52\xec\x29\x4f\xf6\x1b\x97\xb4\x0d\x4f\x66\xf1\x6a\x9d\x2f\xe0\xb2\x2e\x31\x9f\x35\xeb\xfc\x30\xd6\xfe\x39\xe6\xcb\x67\x0d\x5c\x30\xed\x7c\xf3\xca\x49\x7a\x14\xf5\x3d\x3d\x0a\x1e\xaf\xf1\x17\xbc\x97\xb6\x92\x8c\x79\x07\xcc\x97\x36\xc4\x38\x1b\x58\x13\xfe\x5e\x61\x60\x11\x90\x8c\xb4\xe6\x4f\xa4\xe9\x1e\xe0\x7c\xe6\x35\xe6\xb7\xdd\xce\xda\xdf\x32\x19\xb7\x3d\x15\x24\x94\xde\x05\x14\x1a\x9f\x89\x53\xa7\xbf\xc1\x50\x59\xce\x7f\x22\x5f\x3c\x17\x69\xd2\x2b\x5d\x5a\xfb\xa6\xba\xa5\xb5\xe7\x16\xbd\x70\x65\xc5\xef\x51\xf5\xab\xca\x1d\x70\x18\xe9\xde\x28\xa4\xff\x84\xc7\xed\x39\xb4\x34\x6e\x51\xc8\x40\x79\xc7\xf8\xc0\xc7\x7c\x7e\x43\xbe\x2e\x4a\x1b\x37\x0f\xc8\x42\x79\x8d\x69\x4d\xf8\x5b\x96\x5c\x85\x9a\x5a\xbf\x2c\x17\xd7\xab\xaf\xf3\xeb\xf9\x72\x3d\xbf\x9a\x9c\xf4\xd3\xc8\xd2\xd8\x90\x8e\xaf\x91\x77\x5e\x83\xf3\x0c\xa6\x28\x2d\xc6\xdc\xa8\x3f\x8d\x9f\xdc\x48\x9a\x80\xb0\xf4\xbc\x78\xda\x9f\x13\x79\x7a\xc9\xf9\x09\x1b\x69\xed\x07\x30\x07\xc0\x24\x11\x15\xa9\xb5\xe8\x53\x53\xb0\x8f\x9d\x73\x8f\x94\x86\xf6\x55\x06\xcd\x8b\x16\x29\x2a\x95\x28\xea\xc8\x62\x27\x9d\xb6\x48\x01\xa6\xf0\xd7\x11\xfd\xc7\x9d\x26\x5d\x34\x3a\x78\xd0\xdb\x62\xe8\x3f\x6c\xf7\x46\x23\x9a\xf4\x9d\xc0\x0c\x1b\xec\xb0\xdb\x9f\x15\x27\x8d\x2f\x36\xe9\x51\xaf\xbf\x42\xdf\xbe\x23\x1d\x93\xac\x87\x43\xd4\xa8\x77\x01\x41\x5a\xfb\xaf\x10\x68\x14\xef\x6d\x10\x9c\x55\xfb\xdf\xf5\x7d\x1b\x87\x05\xa3\xba\xea\x60\xda\x60\xd0\x23\x8e\xe8\x99\x1d\x83\x30\x1e\x68\xfe\x38\x3b\x45\xbf\xa6\x19\x1a\x92\xe6\x91\xd4\x5d\xd4\x68\xdb\x19\xa5\x69\x2f\x57\x36\x89\x7d\x00\x9f\x61\xbd\x33\x01\x1a\xa1\x36\x01\x4a\x49\x0c\x7e\x03\xd2\xc1\xfc\xd7\xd5\xfc\x66\x11\xfb\xf4\xf2\x2b\x5c\xae\x16\x79\x4f\xce\xdf\x5d\xc6\x7f\x0c\xd1\x4e\x35\xe5\x0d\x84\xbd\x53\xae\x23\x8c\x58\xd2\x16\x7b\xef\x7c\x19\x93\x85\x69\xda\x63\x63\xfb\xa7\x15\x8a\x50\xa3\x63\x23\x6d\x98\x2e\xbd\xeb\x75\x96\x92\xf6\xbb\x26\xbe\x28\x09\x43\x30\xde\x9d\xd8\xfd\x43\x1a\x16\x1b\x4f\x82\x50\xea\xc7\x13\x06\x6c\x0a\xf4\x15\x9f\xd8\x29\x90\xa5\x96\x2c\xeb\xad\x03\x65\x26\xe4\x8a\x5c\x43\x21\x7c\x28\x91\x4c\x8d\xe9\xf1\xd0\xb0\x57\xf6\x16\x93\x33\x86\xa5\xf3\xa6\xa2\x73\x67\x9f\xf6\x3c\xb2\x53\x07\xf0\x32\xf6\xd9\x21\xd4\x59\x0f\xd9\xac\x03\x32\xdb\xe3\x36\x79\x91\x47\xf1\x27\xf8\x41\xa2\xf7\x20\xd1\x70\xf2\x3c\x73\xc4\x7c\xf5\x20\xf9\x7f\x72\xe7\x9f\x00\x00\x00\xff\xff\x58\x7b\xb8\xd5\xe3\x0d\x00\x00")

func bindataExecpb2grpcPyBytes() ([]byte, error) {
	return bindataRead(
		_bindataExecpb2grpcPy,
		"exec_pb2_grpc.py",
	)
}

func bindataExecpb2grpcPy() (*asset, error) {
	bytes, err := bindataExecpb2grpcPyBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name:        "exec_pb2_grpc.py",
		size:        3555,
		md5checksum: "",
		mode:        os.FileMode(420),
		modTime:     time.Unix(1587596493, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataSifterexecPy = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x56\x5d\x6f\xdb\x36\x17\xbe\xd7\xaf\x38\x75\x61\x84\x6e\x5d\xd5\x2d\xde\x77\x17\xc6\x74\x31\xa4\x19\x50\x6c\x48\x82\xa5\x37\x43\x66\x10\x34\x75\xe4\x70\xa1\x48\x95\xa4\x12\x1b\xc3\xfe\xfb\x40\x8a\x94\x25\x7f\x74\x59\x79\x91\xc8\xe2\x73\x3e\x9f\xf3\x50\x7c\xfd\xea\x7d\x6b\xcd\xfb\xb5\x50\xef\x51\x3d\x41\xb3\x73\x0f\x5a\x65\x99\xa8\x1b\x6d\x1c\xd8\x9d\x4d\x8f\x1b\xd3\xf0\xf4\xec\x44\x8d\xe9\xf9\x4f\xab\x55\x7a\x96\x7a\xb3\x11\x6a\xd3\x9b\x3b\xc3\x38\xae\x19\x7f\x4c\x2f\x70\x8b\x9c\x36\xeb\x8f\x87\xbf\x69\xf0\x5e\x19\x5d\x03\xd7\x8a\xb7\xc6\xa0\x72\x10\x41\x55\xeb\x5a\x83\x36\xeb\xf6\x85\xb2\x0d\xf2\x7e\x33\x62\x2b\xc3\x6a\x9c\xc3\x06\xbb\x27\xa1\x2a\x9d\x65\xf4\xe6\xfa\x8a\x7e\xfa\xe9\x77\xfa\xf9\x9a\xde\x5d\x5d\xde\x5c\x7f\xba\x83\x02\x7e\x58\xc0\x9b\xee\xcf\xc7\xff\x65\xd9\xeb\x98\x73\xbe\x66\x56\xf0\x4b\xad\x2a\xb1\x21\x95\x90\xa8\x58\x8d\xc5\x85\x15\x95\x43\xf3\xce\xe7\x99\x4b\xbd\xb9\x98\x83\xc4\x27\x94\x45\xb2\xfa\x7c\xfd\xf3\xcd\x2c\xcb\xb8\x64\xd6\xc2\x25\x93\x92\xad\x25\x5e\xea\x12\x97\x19\x00\x40\x89\x15\x50\x2a\x94\x70\x94\x12\x8b\xb2\x9a\x43\xd5\x2a\x7e\x1d\x92\xe5\xba\xc4\x59\x87\xf3\xcb\x6f\xe7\x3e\x2a\x14\x3d\x68\xbc\xe9\x0d\xa0\x08\x76\xe3\x0d\xcf\x5c\x01\x7f\xfd\xdd\xbf\xf5\xf9\x12\x8f\x9b\xf7\x80\x59\xd6\x67\xc4\x99\x94\x31\x9b\x27\x26\x5b\xb4\x83\x2c\x0c\xba\xd6\xa8\xde\xea\xbe\xcf\x6a\x45\xde\x44\x70\x96\xea\xbd\xdd\xdd\x85\xf6\x5c\x6d\x91\x2f\xb3\xd3\x05\x1f\x16\x18\x6b\x18\xe4\xda\xbf\xa6\xaa\xad\xa1\x80\xc5\x78\xc7\xea\xd6\xf0\x68\xd2\x87\xb8\xd4\x75\x23\x24\xc6\x22\x0c\x7e\x6d\xd1\x3a\xdf\x51\xe5\x70\xeb\x06\x31\x13\x4f\x7e\x20\xc8\x24\x9a\x2d\x61\x6a\x27\xbd\x59\x88\x3d\xeb\x2d\x38\x14\x23\x22\x49\x82\x79\x4e\x9c\xd0\xea\x9c\x61\x5f\xc7\xfd\xa8\xa2\x95\x67\xec\x54\x49\xc7\xb0\xa1\xdf\xde\x42\xb7\x0e\x8a\x5e\x28\x79\x2c\xe1\x37\xb4\xad\x74\x64\x36\x84\xe5\xa2\x84\x62\xdc\xce\x33\x4d\x7e\x5b\xc0\x87\x43\xc6\x75\xeb\x06\xfd\xdd\x4f\xc8\x37\x9a\xcb\x87\xe1\xee\x87\xd9\xaf\x7a\x4c\x2f\x47\x28\x46\xea\x24\x43\xe1\x92\xd9\xbe\x10\x67\x76\xfb\x08\xc7\x14\xfa\xd9\x15\x6a\x03\x53\x0b\x5a\x79\x1e\x61\x0a\x64\x18\x7a\x4f\x4f\xc9\x1c\x1b\x38\x0e\xb5\x31\xc7\xa0\x08\x67\x56\x2e\x35\x2b\x2d\x19\x81\x47\xd8\x30\xed\x9e\xbc\xa0\x03\x9e\x44\x10\x80\xf0\x9a\xe7\x41\x45\xc7\x66\x7a\xc8\xd6\x11\x4d\x01\x91\x0f\xd3\x28\xdb\xba\xb1\x24\x04\x1b\xc3\x12\x2f\x03\x51\x73\x6c\x1c\x5c\x85\x7f\x42\x2b\x60\x16\x70\xf9\x1f\x83\x4b\xe5\x49\xdb\xd9\x1c\xb7\x9c\x86\x96\xce\xee\x3f\xae\x72\xb7\xa6\x0a\xb7\xce\xff\x97\x42\xa1\xd2\xf0\x6e\x30\x22\x81\xec\x4e\xb8\xc3\x01\x1e\x31\x9e\xdb\x46\x0a\x47\x26\x7f\xa8\xc9\x38\x22\x1a\xa3\xcd\xaf\x42\x79\xeb\xc9\x64\xb4\x25\x2a\x9f\xcf\x8f\x20\x51\x91\x83\xd3\xf0\x94\x75\x18\x33\xa9\x56\x07\xdd\x0c\x10\x28\xf6\x5f\x9b\xbc\xd2\xa6\x66\x8e\xe2\x96\x93\x19\xbc\x05\x9f\x13\xbc\xdd\xbb\xfa\xc6\x78\xf9\xa3\xec\xca\xe3\x96\x69\xb6\xa2\xff\xd9\x19\x72\x7a\xc9\xdc\x1a\xcd\xd1\xda\xb1\x6a\xa8\x70\x68\x98\xd3\xe6\x05\x67\x53\x1c\xec\xe8\x47\x9b\x41\x1b\x2b\x6d\xbc\x47\x10\xea\xc8\xf1\xf2\xb0\xa1\x06\xbf\x76\x87\xac\x50\x7b\x71\x1e\xf7\xf5\x48\xbb\x07\xba\x4d\xeb\x48\x8e\xa7\xb3\x3f\x27\xcb\xbd\x24\x4f\xc9\x31\xad\x93\xb2\x3c\x21\xc9\xb4\xf6\xd2\x3c\x23\xc2\xb4\xfe\x5d\x0f\x3d\xf2\x65\xa2\x4c\x6b\x27\x50\x96\x03\x6d\xa6\xf5\x02\x8d\x7e\x4f\x6e\x69\xc4\xad\x33\xe4\x4c\x46\xdf\x37\xc7\x87\xf5\x64\x59\x26\xfc\x17\xdc\x1f\x76\x94\x42\x51\xc0\x84\xd2\x9a\x09\x45\xe9\x24\x7e\xe2\x2d\x9a\x27\xf4\xc9\xf8\x3b\x5b\xde\xfd\x22\xf1\x8e\x96\x7f\x79\x30\xc8\xca\x5b\xad\xa5\xcf\xa0\x75\xda\x90\x9a\x6d\xe9\xb3\x36\x8f\x68\x6c\xf1\x61\x11\xe3\x8f\x2e\x7e\x39\x2b\x4b\x9a\xf0\x77\x68\x9e\x04\x47\x43\x9d\xa6\xd1\x79\xcc\x78\x78\xdd\x20\xb3\x79\x4c\xa4\xf3\x17\xee\x82\x05\xfc\x7f\xb1\x58\x74\xf7\x87\xe7\x07\x21\x11\xbe\x98\x76\xd0\x7d\x85\xcf\x34\x02\x3b\xdb\x10\x58\x28\x8b\xbc\x35\x18\xb6\xc8\xc5\xfd\x72\xb9\x5a\x4e\xed\x05\x4c\x83\xd3\x7d\xbb\x44\xb5\x77\xf0\xaa\x80\xc5\x98\xd5\xb5\x41\xf6\xd8\xbf\x09\xa0\xfe\x2b\x1b\x63\x26\xeb\x41\x17\x73\xeb\x98\x49\xb4\x37\x46\x28\x47\x3c\x62\x0e\x95\x6c\xed\x43\xe1\xd3\xef\xf6\xc6\xec\xde\x75\x14\x04\x63\x2c\xbd\xe2\x42\x88\x69\x19\x98\x0e\x69\xc7\xeb\xde\x48\xbe\xa7\x7a\x12\x40\xa2\xc6\xdc\x4a\xc4\x86\x9c\xb8\x32\x27\xc2\xc2\x60\xff\x82\xbb\xb5\x66\xa6\xfc\xac\x1c\x1a\xd3\x36\x6e\x78\xbd\x8b\x15\xe9\x86\x2c\x66\xd9\x3f\x01\x00\x00\xff\xff\x81\x83\xff\x22\x59\x0c\x00\x00")

func bindataSifterexecPyBytes() ([]byte, error) {
	return bindataRead(
		_bindataSifterexecPy,
		"sifter-exec.py",
	)
}

func bindataSifterexecPy() (*asset, error) {
	bytes, err := bindataSifterexecPyBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name:        "sifter-exec.py",
		size:        3161,
		md5checksum: "",
		mode:        os.FileMode(509),
		modTime:     time.Unix(1607993019, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

//
// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
//
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
}

//
// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
// nolint: deadcode
//
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

//
// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or could not be loaded.
//
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
}

//
// AssetNames returns the names of the assets.
// nolint: deadcode
//
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

//
// _bindata is a table, holding each asset generator, mapped to its name.
//
var _bindata = map[string]func() (*asset, error){
	"exec_pb2.py":      bindataExecpb2Py,
	"exec_pb2_grpc.py": bindataExecpb2grpcPy,
	"sifter-exec.py":   bindataSifterexecPy,
}

//
// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
//
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, &os.PathError{
					Op:   "open",
					Path: name,
					Err:  os.ErrNotExist,
				}
			}
		}
	}
	if node.Func != nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: name,
			Err:  os.ErrNotExist,
		}
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{Func: nil, Children: map[string]*bintree{
	"exec_pb2.py":      {Func: bindataExecpb2Py, Children: map[string]*bintree{}},
	"exec_pb2_grpc.py": {Func: bindataExecpb2grpcPy, Children: map[string]*bintree{}},
	"sifter-exec.py":   {Func: bindataSifterexecPy, Children: map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	return os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
