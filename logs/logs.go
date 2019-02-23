package logs

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	consts "github.com/ihor-sokoliuk/newsbot/configs"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NewsBotLogger represents a logger for News Bot
type NewsBotLogger struct {
	log.Logger
}

// NewLogger creates a default logger
func NewLogger(prefix string) *log.Logger {
	fileRotation := &lumberjack.Logger{
		Filename:   consts.ProjectName + ".log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     3,    //days
		Compress:   true, // disabled by default
	}
	mw := io.MultiWriter(os.Stderr, fileRotation)
	return log.New(mw, prefix, log.Flags())
}

// HandleError logs an error message
func (logger *NewsBotLogger) HandleError(err error, args ...interface{}) (wasError bool) {
	if err != nil {
		// notice that we're using 1, so it will actually log the where
		// the error happened, 0 = this function, we don't want that.
		pc, fn, line, _ := runtime.Caller(1)
		errorMsg := fmt.Sprintf("[ERROR]  %s:%d in %s\t%v.", cutFilePath(fn), line, cutMethodPath(runtime.FuncForPC(pc).Name()), err.Error())
		if len(args) > 0 {
			errorMsg = fmt.Sprintf("%v Args: %v", errorMsg, args)
		}
		logger.Println(errorMsg)
		wasError = true
	}
	return wasError
}

// HandlePanic logs a panic and then panics
func (logger *NewsBotLogger) HandlePanic(err error, args ...interface{}) {
	if err != nil {
		pc, fn, line, _ := runtime.Caller(1)
		logger.Println(fmt.Sprintf("[PANIC]  %s:%d in %s\t%v", cutFilePath(fn), line, cutMethodPath(runtime.FuncForPC(pc).Name()), err.Error()))
		panic(err)
	}
}

// Info logs an info message
func (logger *NewsBotLogger) Info(msg string, args ...interface{}) {
	pc, fn, line, _ := runtime.Caller(1)
	logger.Println(fmt.Sprintf("[INFO]   %s:%d in %s\t%v", cutFilePath(fn), line, cutMethodPath(runtime.FuncForPC(pc).Name()), msg))
}

func cutFilePath(fn string) string {
	i := strings.Index(fn, consts.ProjectName)
	if i > 0 {
		return fn[i:]
	}
	return fn
}

func cutMethodPath(method string) string {
	return method[strings.LastIndex(method, "/")+1:]
}

// getGID retruns goroutine ID
func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}
