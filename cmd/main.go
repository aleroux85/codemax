package main

import (
	"flag"
	"fmt"

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
		return
	}
	fmt.Printf("%+v\n", *lr)
}
