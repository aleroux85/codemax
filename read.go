package codemax

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type logRead struct {
	walk      bool
	locations []string
	files     map[string]filelog
}

type filelog struct {
	history    []commit
	complexity uint
	changefreq uint
	longlines  uint
	numoflines uint
}

type commit struct {
	hash       string
	date       time.Time
	operation  string
	complexity uint
	longlines  uint
	numoflines uint
}

func NewLogRead() *logRead {
	lr := &logRead{}
	lr.files = map[string]filelog{}
	lr.locations = []string{}
	return lr
}

func (lr *logRead) EnableWalk() {
	lr.walk = true
}

func (lr *logRead) SetLocations(ls ...string) {
	lr.locations = ls
}

func (lr logRead) inLocation(fn string) bool {
	if len(lr.locations) == 0 {
		return true
	}
	for _, loc := range lr.locations {
		if strings.HasPrefix(fn, loc) {
			return true
		}
	}
	return false

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

	if lr.walk {
		out, err := exec.Command("git", "checkout", "master").Output()
		if err != nil {
			fmt.Println("error running git checkout,", err, out)
		}
	}

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

	if lr.walk {
		out, err := exec.Command("git", "checkout", info[0]).Output()
		if err != nil {
			fmt.Println("error running git checkout,", err, out)
		}
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

		if !lr.inLocation(info[8]) {
			continue
		}

		c := commit{
			hash:      info[0],
			date:      date,
			operation: info[7],
		}
		if fl, found := lr.files[info[8]]; found {
			fl.history = append(fl.history, c)
			lr.files[info[8]] = fl
		} else {
			lr.files[info[8]] = filelog{
				history: []commit{c},
			}
		}

		err = lr.scan(info[8])
		if err != nil {
			return err
		}
	}

	return nil
}

func (lr *logRead) NumFiles() uint {
	return uint(len(lr.files))
}

func (lr *logRead) scan(fn string) error {
	if !lr.walk {
		return nil
	}
	lastCommit := len(lr.files[fn].history) - 1
	if lr.files[fn].history[lastCommit].operation == "D" {
		return nil
	}

	f, err := os.Open(fn)
	if err != nil {
		fmt.Println("error opening source file,", err)
		return err
	}

	complexity := 0
	longLines := 0
	numberOfLines := 0

	scn := bufio.NewScanner(f)
	for scn.Scan() {
		l := scn.Text()
		ol := len(l)
		tl := len(strings.TrimSpace(l))

		complexity += (ol - tl)

		if ol > 80 {
			longLines++
		}

		numberOfLines++
	}

	if err := scn.Err(); err != nil {
		fmt.Println("error scanning source file,", err)
		return err
	}

	f.Close()

	file := lr.files[fn]
	file.history[lastCommit].complexity = uint(complexity)
	file.history[lastCommit].longlines = uint(longLines)
	file.history[lastCommit].numoflines = uint(numberOfLines)
	lr.files[fn] = file

	return nil
}
