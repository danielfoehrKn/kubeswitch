[![status](https://github.com/t-tomalak/logrus-easy-formatter/workflows/Go/badge.svg)](https://github.com/t-tomalak/logrus-easy-formatter/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/t-tomalak/logrus-easy-formatter)](https://goreportcard.com/report/github.com/t-tomalak/logrus-easy-formatter)
## Logrus Easy Formatter
Provided formatter allow to easily format [Logrus](https://github.com/sirupsen/logrus) log output
Some inspiration taken from [logrus-prefixed-formatter](https://github.com/x-cray/logrus-prefixed-formatter)

## Default output
When format options are not provided `Formatter` will output
```bash
[INFO]: 2006-01-02T15:04:05Z07:00 - Log message
```

## Sample Usage
Sample usage using available option to format output
```go
package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/t-tomalak/logrus-easy-formatter"
)

func main() {
	logger := &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.DebugLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: "2006-01-02 15:04:05",
			LogFormat:       "[%lvl%]: %time% - %msg%",
		},
	}

	logger.Printf("Log message")
}
```
Above sample will produce:
```bash
[INFO]: 27-02-2018 19:16:55 - Log message
```

##### Usage with custom fields
Package also allows to include custom fields and format them(for now only limited to strings)

```go
package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/t-tomalak/logrus-easy-formatter"
)

func main() {
	logger := &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.DebugLevel,
		Formatter: &easy.Formatter{
			LogFormat: "[%lvl%]: %time% - %msg% {%customField%}",
		},
	}

	logger.WithField("customField", "Sample value").Printf("Log message")
}
```
And after running will output
```bash
[INFO]: 27-02-2018 19:16:55 - Log message - {Sample value}
```

## ToDo
- [x] Customizable timestamp formats
- [x] Customizable output formats
- [x] Add tests
- [ ] Support for custom fields other then `string`
- [ ] Tests against all characters

## License
This project is under the MIT License. See the LICENSE file for the full license text.