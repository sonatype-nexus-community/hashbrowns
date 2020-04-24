package types

/*type LogLevel int

const (
	//Info LogLevel = iota
	Info = iota
	Debug
	Trace
)

func (l LogLevel) String() string {
	return [...]string{"Info", "Debug", "Trace"}[l]
}
*/
// TODO figure out how to pass reference to Command for LogLevel type

// Config is basic config for hashbrowns
type Config struct {
	LogLevel    int
	Path        string
	User        string
	Token       string
	Server      string
	Application string
	Stage       string
	MaxRetries  int
}
