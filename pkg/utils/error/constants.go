package errorutils

const (
	// security reserved
	ReservedOffset			= 32
	ReservedMask			= ^int64(0) << ReservedOffset

	// error type bits
	ErrorTypeOffset = 24
	ErrorTypeMask   = ^int64(0) << ErrorTypeOffset

	// error sub type bits
	ErrorSubTypeOffset = 12
	ErrorSubTypeMask   = ^int64(0) << ErrorSubTypeOffset

	DefaultErrorCodeMask = ^int64(0)
)

