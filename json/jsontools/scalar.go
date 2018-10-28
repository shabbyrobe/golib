package jsontools

import (
	"bytes"
	"strconv"
)

type LooseBool bool

func (lb LooseBool) String() string {
	if bool(lb) {
		return "true"
	} else {
		return "false"
	}
}

func (lb *LooseBool) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(b, "\"")
	v, err := strconv.ParseBool(string(b))
	if err != nil {
		return err
	}
	*lb = LooseBool(v)
	return nil
}
