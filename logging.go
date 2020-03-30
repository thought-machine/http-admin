package admin

import (
	"html/template"
	"net/http"

	"gopkg.in/op/go-logging.v1"
)

// LoggerInfo is the interface we need from something that can describe & update log levels.
type LoggerInfo interface{
	// ModuleLevels returns a map of all known modules and their level
	ModuleLevels() map[string]logging.Level
	// SetLevel sets the level of a logging module.
	SetLevel(level logging.Level, module string)
}

var loggerInfo LoggerInfo

// A Logger is the interface we log to.
type Logger interface{
	Debugf(msg string, args ...interface{})
	Infof(msg string, args ...interface{})
	Warningf(msg string, args ...interface{})
	Errorf(msg string, args ...interface{})
}

var noLoggersTemplate = template.Must(template.New("noLoggers").Parse(`
<html>
	<head>
		<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.1.0/css/bootstrap.min.css" integrity="sha384-9gVQ4dYFwwWSjIDZnLEWnxCjeSWFphJiwGPXr1jddIhOegiu1FwO5qRGvFXOdJZ4" crossorigin="anonymous">
		<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.1.0/js/bootstrap.min.js" integrity="sha384-uefMccjFJAIv6A+rW+L4AHf99KvxDjWSu1z9VI8SKNVmz4sk7buKt/6v9KI65qnm" crossorigin="anonymous"></script>
	</head>
	<body>
		<div class="alert alert-danger" role="alert">Logging has not been initialized. Call SetLogger().</div>
	</body>
</html>
`))

var loggersTemplate = template.Must(template.New("loggers").Parse(`
<html>
	<head>
		<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.1.0/css/bootstrap.min.css" integrity="sha384-9gVQ4dYFwwWSjIDZnLEWnxCjeSWFphJiwGPXr1jddIhOegiu1FwO5qRGvFXOdJZ4" crossorigin="anonymous">
		<script src="https://stackpath.bootstrapcdn.com/bootstrap/4.1.0/js/bootstrap.min.js" integrity="sha384-uefMccjFJAIv6A+rW+L4AHf99KvxDjWSu1z9VI8SKNVmz4sk7buKt/6v9KI65qnm" crossorigin="anonymous"></script>
	</head>
	<body>
		<table class="table">
			<thead>
				<tr><th>module</th><th>level</th></tr>
			</thead>
			<tbody>
{{range $module, $moduleLevel := $.ModuleLevels}}
				<tr><td>{{if (eq $module "")}}root{{else}}{{$module}}{{end}}</td>
					<td>
						<form action="/admin/logging" method="POST" style="display: inline"><input type="hidden" name="module" value="{{$module}}" /><input type="hidden" name="level" value="DEBUG" /><input type="submit" class="btn btn{{if (ne $moduleLevel 6)}}-outline{{end}}-secondary" value="ALL"/></form>
{{range $level := $.AllLevels}}
						<form action="/admin/logging" method="POST" style="display: inline"><input type="hidden" name="module" value="{{$module}}" /><input type="hidden" name="level" value="{{ $level.String }}" /><input type="submit" class="btn btn{{if (ne $moduleLevel $level)}}-outline{{end}}-{{index $.Colours $level}}" value="{{$level.String}}"/></form>
{{end}}
						<form action="/admin/logging" method="POST" style="display: inline"><input type="hidden" name="module" value="{{$module}}" /><input type="hidden" name="level" value="-1" /><input type="submit" class="btn btn{{if (ne $moduleLevel -1)}}-outline{{end}}-dark" value="OFF"/></form>
					</td>
				</tr>
{{end}}
			</tbody>
</table>
	</body>
</html>
`))

// LoggingHandler returns the current state of all loggers that we know about.
func LoggingHandler(writer http.ResponseWriter, request *http.Request) {
	if loggerInfo == nil {
		noLoggersTemplate.Execute(writer, nil)
		return
	}
	if err := loggersTemplate.Execute(writer, struct {
		AllLevels    []logging.Level
		Colours      map[logging.Level]string
		ModuleLevels map[string]logging.Level
	}{
		AllLevels: []logging.Level{
			logging.DEBUG,
			logging.INFO,
			logging.NOTICE,
			logging.WARNING,
			logging.ERROR,
			logging.CRITICAL,
		},
		Colours: map[logging.Level]string{
			logging.DEBUG:    "info",
			logging.INFO:     "secondary",
			logging.NOTICE:   "success",
			logging.WARNING:  "warning",
			logging.ERROR:    "danger",
			logging.CRITICAL: "danger",
		},
		ModuleLevels: loggerInfo.ModuleLevels(),
	}); err != nil {
		log.Errorf("%s", err)
	}
}

// UpdateLoggingHandler associates a new log level with a given module. There is no way of telling if the module
// has a custom level or if it is "falling back" to the root module, so this does nothing clever.
func UpdateLoggingHandler(writer http.ResponseWriter, request *http.Request) {
	request.ParseForm()
	module := request.Form.Get("module")
	if level, err := logging.LogLevel(request.Form.Get("level")); err != nil {
		log.Warningf("unable to parse level %s - %s - ignoring", request.Form.Get("level"), err)
	} else {
		log.Debugf("Setting level for %s to %s", module, level)
		loggerInfo.SetLevel(level, module)
	}
	http.Redirect(writer, request, "/admin/logging", http.StatusSeeOther)
}
