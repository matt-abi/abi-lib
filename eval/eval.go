package eval

import (
	"bytes"
	"regexp"
)

var e, _ = regexp.Compile(`\$\{.*?\}`)

func ParseEval(eval string, getValue func(key string) string) string {

	b := bytes.NewBuffer(nil)

	s := []byte(eval)
	n := len(s)

	for n > 0 {

		loc := e.FindIndex(s)

		if loc == nil {
			b.Write(s)
			break
		}

		if loc[0] > 0 {
			b.Write(s[0:loc[0]])
		}

		b.WriteString(getValue(string(s[loc[0]+2 : loc[1]-1])))

		s = s[loc[1]:]
		n = n - loc[1]
	}

	return b.String()
}

func HasEval(eval string) bool {
	return e.MatchString(eval)
}
