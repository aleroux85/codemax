package codemax

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
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

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

func (lr *logRead) Read(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		fmt.Println("error opening log file,", err)
		return err
	}

	loglines, err := lineCounter(f)
	totallines := float64(loglines)
	curline := 0.0
	if err != nil {
		fmt.Println("error counting log file lines,", err)
		return err
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		fmt.Println("error rewinding log file,", err)
		return err
	}

	fmt.Print("Processing...")
	scn := bufio.NewScanner(f)
	for scn.Scan() {
		curline++
		fmt.Printf("\rProcessing... %.1f%%", curline/totallines*100)
		if strings.HasPrefix(scn.Text(), "# ") {
			for {
				err := lr.processCommit(scn, &curline)
				if err != nil && err.Error() == "is commit line" {
					continue
				}
				if err != nil {
					fmt.Println("\rProcessing... done")
					return err
				}
				break
			}
		}
	}
	fmt.Println("\rProcessing... done")

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

func (lr *logRead) processCommit(scn *bufio.Scanner, curline *float64) error {
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
		(*curline)++
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

func (lr *logRead) CCData(f io.Writer, date string) error {
	time := time.Now()

	for filename, file := range lr.files {
		if file.history[0].operation == "D" {
			continue
		}

		for i, commit := range file.history {
			if time.After(commit.date) {
				changefreq := 0
				tprevmonth := time.AddDate(0, -1, 0)
				for j := i; j < len(file.history); j++ {
					if commit.date.Before(tprevmonth) {
						break
					}
					changefreq++
				}
				_, err := fmt.Fprintf(f, "%s,%d,%d,%d,%d\n", filename, commit.complexity, changefreq, commit.longlines, commit.numoflines)
				if err != nil {
					return err
				}
				break
			}
		}
	}

	return nil
}

func (lr *logRead) HistoryData(f io.Writer) error {
	time := time.Now()

	for tstep := 0; tstep < 100; tstep++ {
		time = time.AddDate(0, 0, -7)

		var complexity uint
		var changeFreqAll uint
		var longLines uint
		var numberOfLines uint

		for _, file := range lr.files {
			if file.history[0].operation == "D" {
				continue
			}

			for i, commit := range file.history {
				if time.After(commit.date) {
					var changefreq uint
					tprevmonth := time.AddDate(0, -1, 0)
					for j := i; j < len(file.history); j++ {
						if commit.date.Before(tprevmonth) {
							break
						}
						changefreq++
					}
					complexity += commit.complexity
					changeFreqAll += changefreq
					longLines += commit.longlines
					numberOfLines += commit.numoflines
					break
				}
			}
		}

		_, err := fmt.Fprintf(f, "%s,%d,%d,%d,%d\n", time, complexity, changeFreqAll, longLines, numberOfLines)
		if err != nil {
			return err
		}
	}

	return nil
}
