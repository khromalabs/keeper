package log

import (
    "log"
    "os"
	"io/ioutil"
)

var (
	LogD *log.Logger
	LogI *log.Logger
	LogW *log.Logger
	LogE *log.Logger
)

func init() {
	LogD = log.New(ioutil.Discard, "DEBUG: ", log.Lshortfile)
	LogI = log.New(os.Stdout, "INFO: ", log.Lshortfile)
	LogW = log.New(os.Stdout, "WARNING: ", log.Lshortfile)
	LogE = log.New(os.Stderr, "ERROR: ", log.Lshortfile)
}

func Debug(enable bool) {
	if enable {
		LogD.SetOutput(os.Stdout)
	} else {
		LogD.SetOutput(ioutil.Discard)
	}
}
