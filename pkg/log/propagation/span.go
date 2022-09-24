package propagation

import (
	"strings"
)

func fixB3TID(in string) string {
	if len(in) == 16 {
		in = strings.Repeat("0", 16) + in
	}
	return in
}
