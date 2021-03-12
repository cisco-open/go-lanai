package data

var (
	//Errors
)

type DataError struct {
	error
}

func NewDataError(err error) DataError {
	return DataError{error: err}
}