# Unstructured data wrapper

Companion to reflect.Value, intended for complex deserialisation scenarios (most likely
configuration) where tight control of rich errors is desired and performance is not a
concern.

The main use-case is deserialising JSON or YAML into an `any` (`interface{}`):

```go
var v any
json.Unmarshal(`{"foo": {"bar": [{"baz": "qux"}]}}`, &v)

// Errors are collected in here:
var ctx = &dynamic.ErrContext{}

var uv = dynamic.ValueOf(ctx, "", v)
var baz = uv.Map().Key("foo").Map().Key("bar").Slice().At(0).Key("baz").Str()
fmt.Println(baz) // Outputs 'qux'
```

Once an error is pushed while traversing, the value is marked as 'dead' and
subsequent traversals of the same value do not push further errors:

```go
var v any
json.Unmarshal(`{"foo": {"bar": [{"baz": "qux"}]}}`, &v)
var ctx = &dynamic.ErrContext{}
var uv = dynamic.ValueOf(ctx, "", v)
var baz = uv.
    Map().Key("bork"). // KeyNotFoundError pushed here
    Map().Key("nope"). // No error pushed
    Slice().At(999) // No error pushed

// The dynamic.Value will now be a 'null', which can be checked like so:
if (baz.IsNull()) {
    fmt.Println("no baz found")
    // Attempting to use Str() here would normally push another error, but
    // as 'baz' is a 'dead' value, it does not:
    fmt.Println(baz.Str()) // Outputs ''
} else {
    fmt.Println("unreachable")
}
```


## Expectation Management

This is part of a personal grab-bag of utility code that I add to in a very
ad-hoc fashion. *No API stability guarantees are made*, the code is *not
guaranteed to work*, and anything may be removed at any time as I see fit.

I recommend you copy-paste pieces as-needed (including tests and
license/attribution) into the `internal/` folder of your projects rather than
reference these modules directly as I may change the APIs in here without
warning at any time.

This library will _never_ have a v1.0, so you should hard-pin your version if
you choose not to vendor.

