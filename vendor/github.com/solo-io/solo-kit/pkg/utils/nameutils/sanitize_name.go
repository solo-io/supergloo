package nameutils

import (
	"crypto/md5"
	"fmt"
	"strings"
)

var badChars = []rune{
	'.',
	'_',
}

// sanitize name to make it clean for writing kubernetes objects
func SanitizeName(name string) string {
	name = strings.Map(func(r rune) rune {
		for _, badChar := range badChars {
			if r == badChar {
				return '-'
			}
		}
		return r
	}, name)
	if len(name) > 63 {
		hash := md5.Sum([]byte(name))
		name = fmt.Sprintf("%s-%x", name[:46], hash[:8])
	}
	return name
}
