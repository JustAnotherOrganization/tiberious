package logger

import "log"

/* TODO properly handle errors in a more safe manner and remove the damn
 * brackets from every log message...
 * Implement log to file options and possibly set some stuff to only write
 * during debug mode. */

/*Error logs server-side errors, these are harmful and at least during startup
 * should cause the server to not start. */
func Error(v ...interface{}) {
	log.Fatalln("error:", v)
}

// Info logs standard (non-harmful) information.
func Info(v ...interface{}) {
	log.Println("info:", v)
}
