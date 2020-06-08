package pathtools

import (
	"fmt"
	"path/filepath"
	"strings"
)

func AppendBeforeExt(path string, part string, ext int) (out string, err error) {
	part = filepath.FromSlash(part)
	if strings.IndexRune(part, filepath.Separator) >= 0 {
		return "", fmt.Errorf("pathtools: append part contained path separator")
	}

	base := filepath.Base(path)
	if base == "." || base == "" || base == string(filepath.Separator) {
		return "", fmt.Errorf("pathtools: path contains no base name")
	}

	if ext == 0 {
		return path + part, nil
	}

	end := len(path)
	for i := 0; ext < 0 || i < ext; i++ {
		ext := filepath.Ext(path[:end])
		if ext == "" {
			break
		}
		end -= len(ext)
	}

	if end == len(path) {
		return path + part, nil
	}

	return path[:end] + part + path[end:], nil
}
