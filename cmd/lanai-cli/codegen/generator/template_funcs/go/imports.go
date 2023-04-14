package _go

type Imports []Import

type Import struct {
	Path  string
	Alias string
}

func NewImports() *Imports {
	return &Imports{}
}

func (i *Imports) Add(path string) *Imports {
	*i = append(*i, Import{
		Path: path,
	})
	return i
}

func (i *Imports) AddWithAlias(path, alias string) *Imports {
	*i = append(*i, Import{
		Path:  path,
		Alias: alias,
	})
	return i
}
