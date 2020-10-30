package parser

type PropertyParser func([]byte) (map[string]interface{}, error)
