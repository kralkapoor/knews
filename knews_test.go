package main

import (
	"regexp"
	"testing"
)

func TestThing(t *testing.T) {
	name := "Karl"
	want := regexp.MustCompile(`\b` + name + `\b`)

	otherName := "Karl"

	if !want.MatchString(otherName) {
		t.Error("busted")
	}
}
