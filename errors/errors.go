package errors

import "fmt"

type Error struct {
	Errno  int32  `json:"errno"`
	Errmsg string `json:"errmsg,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Errno, e.Errmsg)
}

func Errorf(errno int32, format string, args ...interface{}) error {
	return &Error{Errno: errno, Errmsg: fmt.Sprintf(format, args...)}
}
