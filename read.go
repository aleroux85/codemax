package codemax

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type logRead struct {
	files map[string]filelog
}

type filelog struct {
	history    []commit
	deleted    bool
	changefreq uint
}

type commit struct {
	hash string
	date time.Time
}

func NewLogRead() *logRead {
	lr := &logRead{}
	lr.files = map[string]filelog{}
	return lr
}

func (lr *logRead) Read(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		fmt.Println("error opening log file,", err)
		return err
	}

	scn := bufio.NewScanner(f)
	for scn.Scan() {
		if strings.HasPrefix(scn.Text(), "# ") {
			for {
				err := lr.processCommit(scn)
				if err != nil && err.Error() == "is commit line" {
					continue
				}
				if err != nil {
					return err
				}
				break
			}
		}
	}

	if err := scn.Err(); err != nil {
		fmt.Println("error scanning log file,", err)
		return err
	}

	f.Close()
	return nil
}

func (lr *logRead) processCommit(scn *bufio.Scanner) error {
	var info [9]string

	l := scn.Text()
	l = strings.TrimPrefix(l, "# ")
	copy(info[:3], strings.Split(l, " - "))

	date, err := time.Parse("2006-01-02 15:04:05 -0700", info[1])
	if err != nil {
		fmt.Println("error formatting date of commit line in log file,", err)
		return err
	}

	for scn.Scan() {
		if strings.HasPrefix(scn.Text(), "# ") {
			return fmt.Errorf("is commit line")
		}
		l := scn.Text()
		if l == "" {
			return nil
		}

		_, err := fmt.Sscanf(l, ":%s %s %s %s %s\t%s", &info[3], &info[4], &info[5], &info[6], &info[7], &info[8])
		if err != nil {
			fmt.Println("error scanning file-change line in log file,", err, "in line", l)
			return err
		}

		c := commit{
			hash: info[0],
			date: date,
		}
		if fl, found := lr.files[info[8]]; found {
			fl.history = append(fl.history, c)
		} else {
			lr.files[info[8]] = filelog{
				history: []commit{c},
			}
		}
	}

	return nil
}

func (lr *logRead) NumFiles() uint {
	return uint(len(lr.files))
}
