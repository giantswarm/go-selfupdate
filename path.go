package selfupdate

import "github.com/giantswarm/go-selfupdate/internal"

func ExecutablePath() (string, error) {
	return internal.GetExecutablePath()
}
