package pkg

import (
	"dario.cat/mergo"
)

func dummy() error {
	return mergo.Merge(map[string]string{}, map[string]string{})
}