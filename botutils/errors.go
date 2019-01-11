package botutils

import (
	"errors"
	"fmt"
)

func HttpStatusCodeError(code int, src string) error {
	msg := fmt.Sprintf("%d received from %s", code, src)
	return errors.New(msg)
}
