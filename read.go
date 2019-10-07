package codemax

import (
	"time"
)

type logRead struct {
	files map[string]file
}

type file struct {
	history []commit
}

type commit struct {
	hash string
	date time.Time
}

func NewLogRead() *logRead {
	lr := &logRead{}
	lr.files = map[string]file{}
	return lr
}

func (lr *logRead) NumFiles() uint {
	return 20
}
