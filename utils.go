package workers

import (
	"reflect"
	"regexp"
	"runtime"
)

var (
	funcNameRegexp      *regexp.Regexp
	funcNameSubexpNames []string
)

func init() {
	funcNameRegexp = regexp.MustCompile("" +
		// package
		"^(?P<package>[^/]*[^.]*)?" +

		".*?" +

		// name
		"(" +
		"(?:glob\\.)?(?P<name>func)(?:\\d+)" + // anonymous func in go >= 1.5 dispatcher.glob.func1 or method.func1
		"|(?P<name>func)(?:·\\d+)" + // anonymous func in go < 1.5, ex. dispatcher.func·002
		"|(?P<name>[^.]+?)(?:\\)[-·]fm)?" + // dispatcher.jobFunc or dispatcher.jobSleepSixSeconds)·fm
		")?$")
	funcNameSubexpNames = funcNameRegexp.SubexpNames()
}

func FunctionName(f interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()

	parts := funcNameRegexp.FindAllStringSubmatch(name, -1)
	if len(parts) > 0 {
		for i, value := range parts[0] {
			switch funcNameSubexpNames[i] {
			case "name":
				if value != "" {
					name += "." + value
				}
			case "package":
				name = value
			}
		}
	}

	return name
}
