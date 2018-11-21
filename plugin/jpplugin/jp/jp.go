package jp

type Dict interface {
	Lookup(string) []DictEntry
	LookupRe(string) ([]DictEntry, error)
}

type DictEntry interface {
	String() string
	Details() string
}
