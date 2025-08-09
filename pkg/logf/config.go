package logf

/* Log config. */
var cfg struct {
	noLog     bool
	outputDir string
}

/* Sets the log config. */
func SetConfig(noLog bool, outputDir string) {
	cfg.noLog = noLog
	cfg.outputDir = outputDir
}
