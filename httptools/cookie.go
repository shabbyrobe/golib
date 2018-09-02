package httptools

import (
	"fmt"
	"net/http"
)

func ParseCookie(raw string) (*http.Cookie, error) {
	header := http.Header{}
	header.Add("Cookie", raw)
	request := http.Request{Header: header}
	cookies := request.Cookies()
	if len(cookies) == 1 {
		return cookies[0], nil
	} else {
		return nil, fmt.Errorf("can not parse raw cookie %q", raw)
	}
}

func ParseCookies(raws ...string) ([]*http.Cookie, error) {
	out := make([]*http.Cookie, len(raws))
	for i, raw := range raws {
		c, err := ParseCookie(raw)
		if err != nil {
			return nil, err
		}
		out[i] = c
	}
	return out, nil
}

func JoinRawCookies(raws ...string) (string, error) {
	parsed, err := ParseCookies(raws...)
	if err != nil {
		return "", err
	}
	return JoinCookies(parsed...)
}

func JoinCookies(cookies ...*http.Cookie) (string, error) {
	hrq, _ := http.NewRequest("GET", "http://example.com", nil)
	for _, cookie := range cookies {
		hrq.AddCookie(cookie)
	}
	return hrq.Header.Get("Cookie"), nil
}
