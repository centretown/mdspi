package main

import (
	"flag"
	"fmt"
	"os"
)

type settingsHelp struct {
	out        string
	in         string
	codec      string
	verbose    string
	serialPort string
	serialBaud string
	webURL     string
}

// Settings - usually from command line
type Settings struct {
	out        string
	in         string
	verbose    bool
	serialPort string
	serialBaud int
	webURL     string
	help       settingsHelp
}

var parsedSettings = Settings{
	help: settingsHelp{
		out:        "output file name",
		in:         "settings input file name",
		verbose:    "detailed messages",
		serialPort: "serial com port name",
		serialBaud: "serial com baud rate",
		webURL:     "web server URL",
	},
	out:        "",
	in:         "settings.json",
	verbose:    false,
	serialPort: "/dev/ttyUSB0",
	serialBaud: 115200,
	webURL:     ":5000",
}

func init() {
	flag.StringVar(&parsedSettings.out, "o",
		parsedSettings.out, parsedSettings.help.out)
	flag.StringVar(&parsedSettings.in, "s",
		parsedSettings.in, parsedSettings.help.in)
	flag.BoolVar(&parsedSettings.verbose, "v",
		parsedSettings.verbose, parsedSettings.help.verbose)
	flag.StringVar(&parsedSettings.serialPort, "sp",
		parsedSettings.serialPort, parsedSettings.help.serialPort)
	flag.IntVar(&parsedSettings.serialBaud, "sb",
		parsedSettings.serialBaud, parsedSettings.help.serialBaud)
	flag.StringVar(&parsedSettings.webURL, "url",
		parsedSettings.webURL, parsedSettings.help.webURL)
}

// MakeSettings - parses and construct settings
func MakeSettings() (s *Settings, err error) {
	s = &parsedSettings
	err = s.Parse()
	return
}

// Parse - parses and returns an error when:
//	input file doesn't exist
func (s *Settings) Parse() (err error) {
	if len(s.in) > 0 {
		_, err = os.Stat(s.in)
		if err != nil {
			err = fmt.Errorf("%v", err)
			return
		}
	}
	return
}
