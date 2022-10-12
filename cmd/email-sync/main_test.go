package main

import (
	"bytes"
	"testing"

	"github.com/ableorg/email-sync/internal/version"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	out := bytes.NewBuffer([]byte{})
	errout := bytes.NewBuffer([]byte{})
	rc := 0
	exit := func(c int) {
		rc = c
	}

	realMain([]string{"secdi",
		"--cfg", "~/.config/secdi/ut_config.yaml",
		"-s", "",
		"version"}, out, errout, exit)
	assert.Equal(t, 1, rc)
	assert.Equal(t, "secdi: error: specify --server flag\n", errout.String())
	assert.Empty(t, out.String())

	out = bytes.NewBuffer([]byte{})
	errout = bytes.NewBuffer([]byte{})
	rc = 0
	realMain([]string{"secdi", "--version"}, out, errout, exit)
	assert.Equal(t, version.Current().String()+"\n", out.String())
	// since our exit func does not call os.Exit, the next parser will fail
	assert.Equal(t, 1, rc)
	assert.NotEmpty(t, errout.String())
}
