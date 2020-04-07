# Development

## Debugging

### Logging
Logging in Service Mesh Hub uses [zap](https://github.com/uber-go/zap). By default the log level will be set to info,
and the encoder will be a JSON encoder. This means that debug level logs will not be shown, and logs will be outputted 
to `os.Stderr` in a machine readable format.

For debugging purposes this can be controlled with the following ENV variables:
1. `LOG_LEVEL`
    * Must be set to a valid `zapcore.Level`. Options can be found [here](https://pkg.go.dev/go.uber.org/zap/zapcore?tab=doc#Level)
    * If none is provided, or the option is invalid, `Info` will be used by default
2. `DEBUG_MODE`
    * If set, the log encoder will be switched to a human readable format. This means it will not be valid JSON anymore.
    More information on how the data will be formatted can be found [here](https://pkg.go.dev/go.uber.org/zap/zapcore?tab=doc#NewConsoleEncoder)
    * Note: If set, this will override the logging level to `debug`, no matter what `LOG_LEVEL` is set to.


### Metrics
metrics currently aren't supported
