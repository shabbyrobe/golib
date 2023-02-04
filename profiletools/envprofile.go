package profiletools

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/pkg/profile"
)

// EnvProfile starts a profile based on environment variables.
//
// Expected env vars are prefixed with envPrefix. Env vars used:
//   - PROFILE: (cpu|block|mem|trace)
//   - PROFILE_PATH: save profiles to this path, instead of OS temp dir
//   - PROFILE_RATE: if using cpu, passed to runtime.SetCPUProfileRate(). if using "block",
//     passed to runtime.SetBlockProfileRate()
//   - PROFILE_QUIET: if not empty, no messages are logged to stderr.
//
// You MUST call Stop() on the returned value when done, even
// if no profile was started:
//
//	defer EnvProfile("MYAPP_").Stop()
//
// --
func EnvProfile(envPrefix string) EnvProfiler {
	var (
		profileEnv = envPrefix + "PROFILE"
		pathEnv    = envPrefix + "PROFILE_PATH"
		rateEnv    = envPrefix + "PROFILE_RATE"
		quietEnv   = envPrefix + "PROFILE_QUIET"

		prof    = os.Getenv(profileEnv)
		path    = os.Getenv(pathEnv)
		rateStr = os.Getenv(rateEnv)
		quiet   = os.Getenv(quietEnv) != ""

		ext = ".pprof"
	)

	var pkind func(*profile.Profile)
	var options []func(p *profile.Profile)
	var rate int64
	var err error

	if rateStr != "" {
		rate, err = strconv.ParseInt(rateStr, 0, 64)
		if err != nil {
			panic(fmt.Errorf("profiletools: profile rate could not be parsed: %v", err))
		}
	}

	var applyRate = func() {}

	switch prof {
	case "cpu":
		pkind = profile.CPUProfile
		if rateStr != "" {
			applyRate = func() {
				runtime.SetCPUProfileRate(int(rate))
			}
		}

	case "block":
		pkind = profile.BlockProfile
		if rateStr != "" {
			applyRate = func() {
				runtime.SetBlockProfileRate(int(rate))
			}
		}

	case "clock":
		pkind = profile.ClockProfile

	case "mem":
		pkind = profile.MemProfile

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

	options = append(options, pkind, profile.ProfilePath(path))
	if quiet {
		options = append(options, profile.Quiet)
	}

	applyRate()
	stop := profile.Start(options...)
	return stopper{
		func() {
			stop.Stop()
			_ = copyFile(expectedFile, lastFile)
			if !quiet {
				log.Printf("profile: %s available at %s\n", prof, lastFile)
			}
		},
	}
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

func copyFile(from, to string) (rerr error) {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()

	st, err := in.Stat()
	if err != nil {
		return err
	}

	expected := st.Size()

	out, err := os.Create(to)
	if err != nil {
		return err
	}

	defer func() {
		cerr := out.Close()
		if rerr == nil && cerr != nil {
			rerr = cerr
		}
	}()

	n, err := io.Copy(out, in)
	if err != nil {
		return err
	}

	if n != expected {
		return fmt.Errorf("iotools: copy expected %d bytes, but only copied %d", expected, n)
	}

	return nil
}
