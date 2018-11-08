package botutils

import "strings"

type Cmd struct {
	Name string
	Args []string
}

func ParseCmd(msg string, prefixes ...string) (c Cmd) {
	msg = strings.TrimSpace(msg)
	for _, p := range prefixes {
		matched := strings.HasPrefix(strings.ToLower(msg),
			strings.ToLower(p))
		if matched {
			// Build and return Cmd
			msg = msg[len(p):]
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
	}
	return 
}
