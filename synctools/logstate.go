// +build ignore

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var (
		file    string
		waiters int
	)

	flag.StringVar(&file, "file", "", "file to parse (- for stdin)")
	flag.IntVar(&waiters, "waiters", 1, "dump waiters. 1 = counts, 2 = full")
	flag.Parse()

	var rdr io.Reader
	if file == "-" {
		rdr = os.Stdin
	} else {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		rdr = f
	}

	var states = make(map[string]*lockState)

	scn := bufio.NewScanner(rdr)
	for scn.Scan() {
		fields := strings.Fields(scn.Text())
		if len(fields) >= 5 {
			lstate := fields[3]

			switch lstate {
			case "lock", "locked", "unlock", "unlocked":
				ptr := fields[1]
				state, ok := states[ptr]
				if !ok {
					state = &lockState{
						ptr:     ptr,
						waiters: map[string]string{},
					}
					states[ptr] = state
				}

				call := fields[2]
				line := strings.Join(fields[4:], " ")

				if lstate == "lock" {
					state.waiters[call] = line
				} else if lstate == "locked" {
					delete(state.waiters, call)
					state.owner = call
				} else if lstate == "unlock" {
					state.owner = ""
				}

				state.at = fields[0]
				state.state = lstate
				state.line = line
			}
		}
	}

	for _, state := range states {
		fmt.Printf("%-12s %-10s %-8s %s\n", state.ptr, state.owner, state.state, state.line)

		if waiters == 1 {
			wlen := len(state.waiters)
			if wlen > 0 {
				cnt := make(map[string]int, wlen)
				for _, w := range state.waiters {
					cnt[w]++
				}
				for l, c := range cnt {
					fmt.Printf("  %-10d %s\n", c, l)
				}
				fmt.Println()
			}

		} else if waiters == 2 {
			for c, w := range state.waiters {
				fmt.Printf("  %-10s %s\n", c, w)
			}
			fmt.Println()
		}
	}

	return nil
}

type lockState struct {
	ptr     string
	state   string
	at      string
	line    string
	owner   string
	waiters map[string]string
}
