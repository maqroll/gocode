package clickhouse

import (
	"testing"
)

func TestCaptureErrMissingCommand(t *testing.T) {
	cli := []string{"i", "do", "not", "exist"}
	ch := &clickhouseType{
		cli: cli,
	}

	expected := "Failed to start process"
	res := ch.captureErr("any")
	if res != expected {
		t.Errorf("error didn't match. Expected:%q, got:%q", expected, res)
	}
}

func TestCaptureErrBadResult(t *testing.T) {
	cli := []string{"go", "no", "exist"}
	ch := &clickhouseType{
		cli: cli,
	}

	noExpected := "Unspecified problem running command"
	res := ch.captureErr("any")
	if res == noExpected {
		t.Errorf("error match no expected result. No expected:%q, got:%q", noExpected, res)
	}
}

func TestCaptureErrAnotherBadResult(t *testing.T) {
	cli := []string{"go", "build"}
	ch := &clickhouseType{
		cli: cli,
	}

	noExpected := "Unspecified problem running command"
	res := ch.captureErr("no.go")
	if res == noExpected {
		t.Errorf("error match no expected result. No expected:%q, got:%q", noExpected, res)
	}
}
