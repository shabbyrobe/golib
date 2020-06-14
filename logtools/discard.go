package logtools

type discard struct{}

var discardLogger Logger = &discard{}

func Discard() Logger {
	return discardLogger
}

func (d *discard) Print(...interface{})          {}
func (d *discard) Printf(string, ...interface{}) {}
func (d *discard) Println(...interface{})        {}

func (d *discard) Fatal(...interface{})          {}
func (d *discard) Fatalf(string, ...interface{}) {}
func (d *discard) Fatalln(...interface{})        {}

func (d *discard) Panic(...interface{})          {}
func (d *discard) Panicf(string, ...interface{}) {}
func (d *discard) Panicln(...interface{})        {}
