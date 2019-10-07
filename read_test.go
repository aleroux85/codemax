package codemax_test

import (
	"strconv"
	"testing"

	"github.com/aleroux85/codemax"
)

func Test(t *testing.T) {
	lr := codemax.NewLogRead()

	exp := "20"
	got := strconv.Itoa(int(lr.NumFiles()))
	if exp != got {
		t.Errorf(`expected %s, got %s`, exp, got)
	}
}
