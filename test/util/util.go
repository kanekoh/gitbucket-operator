package util

import (
	"testing"
)

func NotToHaveError(t *testing.T, err error, message string) {
	if err != nil {
		t.Fatalf(message, err)
	}
}