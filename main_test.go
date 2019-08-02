package main

import (
	"testing"
)

func TestMakeChapterFile(t *testing.T) {
	if err := makeChapterFile("sample.vdr", "sample.chp"); err != nil {
		t.Error(err)
	}
}
