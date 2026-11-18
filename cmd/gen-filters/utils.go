package main

import "unicode"

//functional map func for converting between types. Similar to js version.
func Map[T any, U any](items []T, mapper func(T) U) []U {
	var val []U = make([]U, len(items))
	for i, item := range items {
		val[i] = mapper(item)
	}
	return val
}

func GetTypeName(t string) string {
	if t[0] == '*' {
		return t[1:]
	}
	return t
}

func Pascal(str string) string {
	if str == "" {
		return str
	}
	runes := []rune(str)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func Camel(str string) string {
	if str == "" {
		return str
	}
	runes := []rune(str)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
