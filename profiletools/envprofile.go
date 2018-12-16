package profiletools

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/profile"
	"github.com/shabbyrobe/golib/iotools"
)

// EnvProfile starts a profile based on environment variables.
//
// You MUST call Stop() on the returned value when done, even
// if no profile was started:
//	defer EnvProfile("MYAPP_").Stop()
//
func EnvProfile(envPrefix string) EnvProfiler {
	var (
		profileEnv = envPrefix + "PROFILE"
		pathEnv    = envPrefix + "PROFILE_PATH"
		quietEnv   = envPrefix + "PROFILE_QUIET"

		prof  = os.Getenv(profileEnv)
		path  = os.Getenv(pathEnv)
		quiet = os.Getenv(quietEnv) != ""

		ext = ".pprof"
	)

	var pkind func(*profile.Profile)
	switch prof {
	case "cpu":
		pkind = profile.CPUProfile
	case "mem":
		pkind = profile.MemProfile
	case "block":
		pkind = profile.BlockProfile
	case "trace":
		pkind = profile.TraceProfile
		ext = ".out"
	default:
		return stopper{}
	}

	prog := filepath.Base(os.Args[0])

	if path == "" {
		path = filepath.Join(os.TempDir(), fmt.Sprintf("%s-%d", prog, time.Now().UnixNano()))
	}

	// WARNING: this relies on assumptions about the internals of the profile
	// package, which could change without warning.
	expectedFile := filepath.Join(path, prof+ext)
	lastFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s%s", prog, prof, ext))

	options := []func(p *profile.Profile){pkind, profile.ProfilePath(path)}
	if quiet {
		options = append(options, profile.Quiet)
	}

	stop := profile.Start(options...)
	return stopper{
		func() {
			stop.Stop()
			iotools.CopyFile(expectedFile, lastFile)
			if !quiet {
				log.Printf("profile: %s available at %s\n", prof, lastFile)
			}
		},
	}
	return stop
}

type EnvProfiler interface {
	Stop()
}

type stopper struct {
	stop func()
}

func (d stopper) Stop() {
	if d.stop != nil {
		d.stop()
	}
}
