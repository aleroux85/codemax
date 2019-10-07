package codemax_test

import (
	"strconv"
	"testing"

	"github.com/aleroux85/codemax"
)

func Test(t *testing.T) {
	lr := codemax.NewLogRead()
	err := lr.Read("testing/githist.log")
	if err != nil {
		t.Error(err)
		t.SkipNow()
	}

	exp := "1"
	got := strconv.Itoa(int(lr.NumFiles()))
	if exp != got {
		t.Errorf(`expected %s, got %s`, exp, got)
	}
}

func TestFileOpenError(t *testing.T) {
	lr := codemax.NewLogRead()
	err := lr.Read("testing/githist_none.log")
	if err == nil {
		t.Error("expected error")
	}
}

func TestDateFormatError(t *testing.T) {
	lr := codemax.NewLogRead()
	err := lr.Read("testing/githist-baddate.log")
	if err == nil {
		t.Error("expected error")
	}
}

func TestFileChangeFormatError(t *testing.T) {
	lr := codemax.NewLogRead()
	err := lr.Read("testing/githist-badfilechange.log")
	if err == nil {
		t.Error("expected error")
	}
}
