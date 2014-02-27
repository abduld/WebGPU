package config

var (
	TraceLogOutput string
	InfoLogOutput  string
	WarnLogOutput  string
	ErrorLogOutput string
)

func InitLoggerConfig() {
	conf := NestedRevelConfig

	TraceLogOutput = conf.StringDefault("app.trace", "stdout")
	InfoLogOutput = conf.StringDefault("app.info", "stdout")
	WarnLogOutput = conf.StringDefault("app.warn", "stdout")
	ErrorLogOutput = conf.StringDefault("app.error", "stdout")
}
