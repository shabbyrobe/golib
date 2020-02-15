package interval

import (
	"flag"
	"testing"

	"github.com/shabbyrobe/golib/assert"
)

func TestFlag(t *testing.T) {
	tt := assert.WrapTB(t)

	var intvlFlag FlagVar
	fs := flag.NewFlagSet("", 0)
	fs.Var(&intvlFlag, "intvl", "Interval!")
	tt.MustOK(fs.Parse([]string{"-intvl", "1min"}))

	intvl := intvlFlag.Interval()
	tt.MustEqual(Of1Minute, intvl)
}
