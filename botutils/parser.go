package botutils

import "strings"

type Cmd struct {
	Name string
	Args []string
}

func ParseCmd(msg string, prefix string) (c Cmd) {

	msg = strings.TrimSpace(msg)
	if !strings.HasPrefix(strings.ToLower(msg),
		strings.ToLower(prefix)) {
		return
	}

	//Trim prefix
	msg = msg[len(prefix):]
	f := strings.Fields(msg)

	length := len(msg)
	if length > 1 {
		c.Args = f[1:]
	}
	if length > 0 {
		c.Name = f[0]
	}
	return c
}
