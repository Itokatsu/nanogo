package web

import "strings"

type ResultEntry struct {
	Header  string
	Content []string
}

func (e ResultEntry) Print() string {
	str := e.Header + "\n\t"
	str += strings.Join(e.Content, "\n\t")
	return str
}
