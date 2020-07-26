package readers

import (
  "strings"
)

type CSVParse struct {
	Comma   string
	Comment string
}

func (c *CSVParse) Parse(line string) []string {
	comma := c.Comma
	if comma == "" {
		comma = ","
	}
	comment := c.Comment
	if comment != "" {
		line = strings.Split(line, comment)[0]
	}
	return strings.Split(line, comma)
}


type CSVReader struct {
  Comma   string
  Comment string
}

func (c *CSVReader) Read(in chan []byte) chan []string {
  t := make(chan string, 100)
  go func() {
    defer close(t)
    for i := range in {
      t <- string(i)
    }
  }()
  return c.ReadString(t)
}

func (c *CSVReader) ReadString(in chan string) chan []string {
  comma := c.Comma
  if comma == "" {
    comma = ","
  }
  comment := c.Comment

  out := make(chan []string, 100)
  go func() {
    defer close(out)
    for line := range in {
      if comment != "" {
        line = strings.Split(line, comment)[0]
      }
      if len(line) > 0 {
        out <- strings.Split(line, comma)
      }
    }
  }()
  return out
}
