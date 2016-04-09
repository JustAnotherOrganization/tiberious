package logger

import (
	"io"
	"log"
	"os"

	"tiberious/settings"
)

var (
	config settings.Config
	// Use separate loggers so that errors can be more easily tracked
	errors *log.Logger
	info   *log.Logger
)

func init() {
	config = settings.GetConfig()

	var errwriter io.Writer
	var infowriter io.Writer

	if config.ErrorLog != "" {
		var err error
		errwriter, err = os.OpenFile(config.ErrorLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		errwriter = os.Stderr
	}

	if config.DebugLog != "" {
		var err error
		infowriter, err = os.OpenFile(config.DebugLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		infowriter = os.Stdout
	}

	// TODO Find a way to remove the brackets around all the data?
	errors = log.New(errwriter, "", log.LstdFlags)
	info = log.New(infowriter, "", log.LstdFlags)
}

/*Error logs server-side errors, these are harmful and at least during startup
 * should cause the server to not start. */
func Error(v ...interface{}) {
	// TODO possibly make this non-fatal...
	errors.Fatalln("error: ", v)
}

// Info logs standard (non-harmful) information.
func Info(v ...interface{}) {
	info.Println("info:", v)
}
