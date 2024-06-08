package logtools

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"testing"
	"testing/slogtest"
)

type mapHandlerDriver struct {
	result map[string]any
}

func (d *mapHandlerDriver) NewHandler(t *testing.T) slog.Handler {
	w := io.Discard
	if debug, _ := strconv.ParseBool(os.Getenv("MAP_HANDLER_DEBUG")); debug {
		w = os.Stdout
	}

	h := slog.NewJSONHandler(w, nil)
	return NewMapHandler(slog.LevelDebug, h, func(m map[string]any) error {
		d.result = m
		return nil
	})
}

func (d *mapHandlerDriver) Result(t *testing.T) map[string]any {
	if debug, _ := strconv.ParseBool(os.Getenv("MAP_HANDLER_DEBUG")); debug {
		bts, err := json.Marshal(d.result)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(bts))
	}
	return d.result
}

func TestMapHandler(t *testing.T) {
	d := &mapHandlerDriver{}
	slogtest.Run(t, d.NewHandler, d.Result)
}
