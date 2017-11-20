package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

// appFs is an aero filesystem.  It is stored in a variable so that we can replace it
// with in-memory filesystems for unit tests.
var appFs = afero.NewOsFs()

type logger struct {
	cmdLogger    fileLogger
	testerLogger fileLogger
	daemonLogger fileLogger
}

// Create a new logger that will log in the proper directory.
// Also initializes all necessary directories and files.
func newLogger() (logger, error) {
	cmdLoggerPath := filepath.Join(os.Getenv("WORKSPACE"), "commandOutputs.log")
	cmdLoggerFile, err := os.Create(cmdLoggerPath)
	if err != nil {
		return logger{}, err
	}

	daemonLoggerPath := filepath.Join(os.Getenv("WORKSPACE"), "daemonOutput.log")
	daemonLoggerFile, err := os.Create(daemonLoggerPath)
	if err != nil {
		return logger{}, err
	}

	return logger{
		testerLogger: fileLogger{os.Stdout},
		cmdLogger:    fileLogger{cmdLoggerFile},
		daemonLogger: fileLogger{daemonLoggerFile},
	}, nil
}

type fileLogger struct {
	out io.Writer
}

func (l fileLogger) infoln(msg string) {
	timestamp := time.Now().Format("[15:04:05] ")
	l.println("\n" + timestamp + "=== " + msg + " ===")
}

func (l fileLogger) errorln(msg string) {
	l.println("\n=== Error Text ===\n" + msg + "\n")
}

func (l fileLogger) println(msg string) {
	fmt.Fprintln(l.out, msg)
}

func fileContents(file string) (string, error) {
	a := afero.Afero{
		Fs: appFs,
	}
	contents, err := a.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}
