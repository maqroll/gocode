package clickhouse

const (
	// StopOption stop processing if detects pending tables
	StopOption = "STOP"
	// ProcessOption complete processing if detects pending tables
	ProcessOption = "PROCESS"
	// DeleteOption delete pending tables
	DeleteOption = "DELETE"
)
