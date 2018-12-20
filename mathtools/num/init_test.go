package num

import (
	"flag"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"
)

var (
	fuzzIterations  = fuzzDefaultIterations
	fuzzOpsActive   = allFuzzOps
	fuzzTypesActive = allFuzzTypes
	fuzzSeed        int64
)

func TestMain(m *testing.M) {
	var ops StringList
	var types StringList

	flag.IntVar(&fuzzIterations, "num.fuzziter", fuzzIterations, "Number of iterations to fuzz each op")
	flag.Int64Var(&fuzzSeed, "num.fuzzseed", fuzzSeed, "Seed the RNG")
	flag.Var(&ops, "num.fuzzop", "Fuzz op to run (can pass multiple)")
	flag.Var(&types, "num.fuzztype", "Fuzz type (u128, i128) (can pass multiple)")
	flag.Parse()

	if len(ops) > 0 {
		fuzzOpsActive = nil
		for _, op := range ops {
			fuzzOpsActive = append(fuzzOpsActive, fuzzOp(op))
		}
	}

	if len(types) > 0 {
		fuzzTypesActive = nil
		for _, t := range types {
			fuzzTypesActive = append(fuzzTypesActive, fuzzType(t))
		}
	}

	log.Println("active ops:", fuzzOpsActive)
	log.Println("iterations:", fuzzIterations)

	code := m.Run()
	os.Exit(code)
}

var trimFloatPattern = regexp.MustCompile(`(\.0+$|(\.\d+[1-9])\0+$)`)

func cleanFloatStr(str string) string {
	return trimFloatPattern.ReplaceAllString(str, "$2")
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
