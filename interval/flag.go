package interval

import (
	"flag"
)

// FlagVar wraps an Interval for use with flag.Var:
//
//	var intvlFlag interval.FlagVar
//	fs := flag.NewFlagSet("", 0)
//	fs.Var(&intvlFlag, "intvl", "Interval!")
//	intvl := intvlFlag.Interval()
//
type FlagVar Interval

var _ flag.Value = new(FlagVar)

func (iv FlagVar) IsZero() bool       { return Interval(iv).IsZero() }
func (iv FlagVar) String() string     { return iv.Interval().String() }
func (iv FlagVar) Interval() Interval { return Interval(iv) }
func (iv FlagVar) Get() interface{}   { return iv.Interval() }

func (iv *FlagVar) Set(s string) error {
	if s == "" {
		return nil
	}
	t, err := Parse(s)
	if err != nil {
		return err
	}
	*iv = FlagVar(t)
	return nil
}
