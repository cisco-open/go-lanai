package _go

type Imports []Import

type Import struct {
	Path  string
	Alias string
}

func NewImports() *Imports {
	return &Imports{}
}

// Add only first element in args is used as alias if not empty
func (i *Imports) Add(path string, args...string) *Imports {
	alias := ""
	if len(args) > 0 {
		alias = args[0]
	}
	*i = append(*i, Import{
		Path: path,
		Alias: alias,
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
