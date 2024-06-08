package logtools

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
)

func attrsToMap(attrs ...slog.Attr) map[string]any {
	m := make(map[string]any, len(attrs))
	for _, attr := range attrs {
		setAttr(m, attr)
	}
	return m
}

func setAttr(m map[string]any, attr slog.Attr) {
	if attr.Key == "" {
		if attr.Value.Kind() == slog.KindGroup {
			for _, gattr := range attr.Value.Group() {
				setAttr(m, gattr)
			}
		}
		return
	}

	switch attr.Value.Kind() {
	case slog.KindGroup:
		gattrs := attr.Value.Group()
		if len(gattrs) > 0 {
			m[attr.Key] = (map[string]any)(attrsToMap(gattrs...))
		}
	case slog.KindLogValuer:
		m[attr.Key] = attr.Value.Resolve().Any()
	case slog.KindAny:
		m[attr.Key] = fmt.Sprint(attr.Value)
	default:
		m[attr.Key] = attr.Value.Any()
	}
}

type attrs struct {
	group string
	attrs []slog.Attr
}

type MapHandler struct {
	level  slog.Level
	handle func(map[string]any) error
	next   slog.Handler
	stack  []attrs
}

var _ slog.Handler = (*MapHandler)(nil)

func NewMapHandler(level slog.Level, next slog.Handler, handle func(map[string]any) error) *MapHandler {
	h := &MapHandler{
		level:  level,
		next:   next,
		stack:  []attrs{{}},
		handle: handle,
	}
	return h
}

func (h *MapHandler) Enabled(ctx context.Context, level slog.Level) (enabled bool) {
	if h.next != nil {
		if h.next.Enabled(ctx, level) {
			return true
		}
	}
	if h.handle != nil && level >= h.level {
		return true
	}
	return false
}

func (h *MapHandler) Handle(ctx context.Context, record slog.Record) (rerr error) {
	if h.handle != nil && record.Level >= h.level {
		// First item in the stack is the root, should always create a map as
		// this map has the special attributes attached:
		var current = attrsToMap(h.stack[0].attrs...)
		var entry = current

		// Look backwards to find the last item in the stack with attrs. This helps us
		// avoid adding empty intermediate groups that ultimately have no attrs if we
		// don't have to (this only matters if the record itself has no attrs):
		var end = len(h.stack)
		if record.NumAttrs() == 0 {
			for end >= 1 {
				if len(h.stack[end-1].attrs) > 0 {
					break
				}
				end--
			}
		}

		// For each stack entry after the 0th, up to and including the last group
		// that actually has attrs, build the tree:
		for _, l := range h.stack[1:max(1, end)] {
			next := attrsToMap(l.attrs...)
			current[l.group] = next
			current = next
		}

		// Assign the attrs from the record:
		record.Attrs(func(attr slog.Attr) bool {
			setAttr(current, attr)
			return true
		})

		// Add special attributes to the topmost map last to avid clobbering:
		if !record.Time.IsZero() {
			entry["time"] = record.Time
		}
		entry["level"] = record.Level
		entry["msg"] = record.Message

		rerr = h.handle(entry)
	}

	if h.next != nil {
		if herr := h.next.Handle(ctx, record); herr != nil {
			rerr = errors.Join(rerr, herr)
		}
	}

	return rerr
}

func (h *MapHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	with := *h
	if h.handle != nil {
		with.stack = slices.Clone(with.stack)
		with.stack[len(with.stack)-1].attrs = append(with.stack[len(with.stack)-1].attrs, attrs...)
	}
	if with.next != nil {
		with.next = with.next.WithAttrs(attrs)
	}
	return &with
}

func (h *MapHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	with := *h
	if h.handle != nil {
		with.stack = slices.Clone(with.stack)
		with.stack = append(with.stack, attrs{group: name})
	}
	if with.next != nil {
		with.next = with.next.WithGroup(name)
	}
	return &with
}
