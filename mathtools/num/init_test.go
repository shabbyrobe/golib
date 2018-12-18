package num

import (
	"flag"
	"log"
	"os"
	"strings"
	"testing"
)

var (
	fuzzIterations = fuzzDefaultIterations
	fuzzOpsActive  = allFuzzOps
	fuzzSeed       int64
)

func TestMain(m *testing.M) {
	var ops StringList

	flag.IntVar(&fuzzIterations, "num.fuzziter", fuzzIterations, "Number of iterations to fuzz each op")
	flag.Int64Var(&fuzzSeed, "num.fuzzseed", fuzzSeed, "Seed the RNG")
	flag.Var(&ops, "num.fuzzop", "Fuzz op to run (can pass multiple)")
	flag.Parse()

	if len(ops) > 0 {
		fuzzOpsActive = nil
		for _, op := range ops {
			fuzzOpsActive = append(fuzzOpsActive, fuzzOp(op))
		}
	}

	log.Println("active ops:", fuzzOpsActive)
	log.Println("iterations:", fuzzIterations)

	code := m.Run()
	os.Exit(code)
}

type StringList []string

func (s StringList) Strings() []string { return s }

func (s *StringList) String() string {
	if s == nil {
		return ""
	}
	return strings.Join(*s, ",")
}

func (s *StringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}
