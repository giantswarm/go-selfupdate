package internal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/giantswarm/go-selfupdate/internal"
)

func TestGetExecutablePath(t *testing.T) {
	t.Parallel()

	exe, err := internal.GetExecutablePath()
	assert.NoError(t, err)
	assert.NotEmpty(t, exe)
}
