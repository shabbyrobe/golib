package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/shabbyrobe/golib/unidecode"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	bts, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	into := make([]byte, unidecode.MaxDecodedLen(len(bts)))
	out := unidecode.Decode(bts, into)
	os.Stdout.Write(out)

	return nil
}
