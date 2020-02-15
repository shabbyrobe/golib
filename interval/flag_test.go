package interval

import (
	"flag"
	"testing"
)

func TestFlag(t *testing.T) {
	var intvlFlag FlagVar
	fs := flag.NewFlagSet("", 0)
	fs.Var(&intvlFlag, "intvl", "Interval!")
	if err := fs.Parse([]string{"-intvl", "1min"}); err != nil {
		t.Fatal(err)
	}

	intvl := intvlFlag.Interval()
	if intvl != Of1Minute {
		t.Fatal(intvl, "!=", Of1Minute)
	}
}
