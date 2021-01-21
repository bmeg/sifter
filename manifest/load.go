package manifest

import (
  "os"
  "crypto/md5"
  "fmt"
  "io"
  "io/ioutil"
  "path/filepath"
  "github.com/ghodss/yaml"
)

type Entry struct {
  Path      string  `json:"path" jsonschema_description:"relative storage path"`
  MD5       string  `json:"md5" jsonschema_description:"MD5 of file"`
  Source    string  `json:"source" jsonschema_description:"URL of original download"`
  Timestamp string  `json:"timestamp" jsonschema_description:"timestamp of file"`
  realPath  string
}

type File struct {
  Entries   []Entry
  path      string
}


func Load(relpath string) (File, error) {
	// Try to get absolute path. If it fails, fall back to relative path.
	path, abserr := filepath.Abs(relpath)
	if abserr != nil {
		path = relpath
	}

	// Read file
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return File{}, fmt.Errorf("failed to read config at path %s: \n%v", path, err)
	}

  entlist := []Entry{}
  err = yaml.Unmarshal(source, &entlist)
  if err != nil {
		return File{}, fmt.Errorf("failed to parse config at path %s: \n%v", path, err)
	}
  baseDir := filepath.Dir(path)
  for i := range entlist {
    entlist[i].realPath = filepath.Join(baseDir, entlist[i].Path)
  }
  return File{entlist, path}, nil
}

// ParseDataFile parses input file
func ParseDataFile(path string, data *map[string]interface{}) error {

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read data at path %s: \n%v", path, err)
	}
	return yaml.Unmarshal(raw, data)
}



func (e Entry) Exists() bool {
  _, err := os.Stat(e.realPath)
  if os.IsNotExist(err) {
      return false
  }
  return true
}



func (e Entry)CalcMD5() (string, error) {
  f, err := os.Open(e.realPath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
