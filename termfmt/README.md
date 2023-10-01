# Terminal formattimg & styling

Copy-pasta for terminal formatting, supporting `fmt.Sprintf`. Not built for
speed or for minimising output, built for copy-pastability and simplicity.

```go
// Somewhere global or shared, probably:
var failed = termfmt.Fg(255, 0, 0, 96, termfmt.Red)
var succeeded = termfmt.Fg(255, 0, 0, 96, termfmt.Red)
var count = termfmt.FgRGB(255, 255, 255).Italic()
var link = termfmt.FgRGB(192, 192, 192).Bold()

// To use:
if borked {
    fmt.Sprintf("%s (%d): %s",
        failed.V("failed"),
        count.V(cnt),
        link.Linked(statusURL).V("more info"))
} else {
    fmt.Sprintf("%s (%d): All good mate!",
        failed.V("failed"),
        count.V(cnt))
}
```

It tends to do what you expect when you use field widths. This will show three
fields padded to 10 characters, separated by a single space, without the format
characters counting towards the widths:

```go
var (
    good   = FgRGB(0, 255, 0)
    bad    = FgRGB(255, 0, 0)
    linked = Linked("http://example.com")
)

fmt.Printf("%-10s %-10s %-10s",
    good.V("yes"),
    bad.V("no"),
    linked.V("click me"))
```

## Formatters:

- `Linked(url)`: Uses OSC 8 if supported to provide a hyperlink
- `Bold()`: Yep
- `Italic()`: That too
- `Fg(r, g, b, c256, c16)`: Foreground colour, requires RGB, 256 colour and 16 colour values.
- `Bg(r, g, b, c256, c16)`: Background colour, as above
- `FgRGB(r, g, b)`: RGB only, doesn't bother with fallbacks.

## Color support:

RGB is enabled by default (it's 2023). If you want to use downgraded colours
(provided you set them in your styles), call `termfmt.RGBSupported(false)` and
`termfmt.C256Supported(false)` to use the fallbacks.

## Expectation Management

This is part of a personal grab-bag of utility code that I add to in a very
ad-hoc fashion. *No API stability guarantees are made*, the code is *not
guaranteed to work*, and anything may be removed at any time as I see fit.

I recommend you copy-paste pieces as-needed into the `internal/` folder of your
projects rather than reference these modules directly as I may change the APIs
in here without warning at any time. If you need me to disavow copyright I will,
but this stuff is not novel and shouldn't be bound by any.
