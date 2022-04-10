package dynamic

import "fmt"

func pathWithIdx(p string, idx int) string {
	return fmt.Sprintf("%s/%d", p, idx)
}

func pathWithKey(p string, key string) string {
	return fmt.Sprintf("%s/%s", p, key)
}

type value interface {
	IsNull() bool
	IsValid() bool
	Kind() Kind
	Path() string
	Unwrap() any
	Descend(part ...any) Value
	TryDescend(part ...any) Value
	Reject(err error)
}
