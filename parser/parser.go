package parser

import "strings"

type ParsedCmd struct {
	Name string
	Args []string
}

func Cmd(msg string, prefix string) (parsed ParsedCmd) {

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
		parsed.Args = f[1:]
	}
	if length > 0 {
		parsed.Name = f[0]
	}
	return parsed
}

/*
func Date() {

} */
