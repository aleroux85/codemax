package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/aleroux85/codemax"
)

func main() {
	flag.Parse()
	locations := flag.Args()

	lr := codemax.NewLogRead()
	lr.EnableWalk()
	lr.SetLocations(locations...)

	err := lr.Read("githist.log")
	if err != nil {
		fmt.Println(err)
		return
	}

	f, err := os.Create("file-report.csv")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	fmt.Fprintln(f, "file, complexity, changeFreqAll, longLines, numberOfLines")
	lr.CCData(f, "now")

	f, err = os.Create("history-report.csv")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	fmt.Fprintln(f, "time, complexity, changeFreqAll, longLines, numberOfLines")
	lr.HistoryData(f)
}
