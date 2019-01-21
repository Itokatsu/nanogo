package web

import (
	"fmt"
	"strings"
)

// CJK space
const offset = "\u3000\u3000"

type ResultEntry struct {
	Header  string
	Content []string
}

func (e ResultEntry) String() string {
	contentText := strings.Join(e.Content, "\n")
	return fmt.Sprintf(":small_orange_diamond: %s\n%s", e.Header, contentText)
}
