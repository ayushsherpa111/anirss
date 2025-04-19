package objects

type Logging struct {
	Message string
	Error   error
	Level   int
}

const (
	L_INFO = iota
	L_ERROR
)
