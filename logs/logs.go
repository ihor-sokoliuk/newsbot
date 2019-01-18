package logs

import (
	"fmt"
	consts "github.com/ihor-sokoliuk/newsbot/configs"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
)

type NewsBotLogger struct {
	log.Logger
}

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

func (logger *NewsBotLogger) HandleError(err error, args ...interface{}) (wasError bool) {
	if err != nil {
		// notice that we're using 1, so it will actually log the where
		// the error happened, 0 = this function, we don't want that.
		pc, fn, line, _ := runtime.Caller(1)
		errorMsg := fmt.Sprintf("%9v\t%s:%d in %s\t%v.", "[ERROR]", cutFilePath(fn), line, cutMethodPath(runtime.FuncForPC(pc).Name()), err.Error())
		if len(args) > 0 {
			errorMsg = fmt.Sprintf("%v Args: %v", errorMsg, args)
		}
		logger.Println(errorMsg)
		wasError = true
	}
	return wasError
}

func (logger *NewsBotLogger) HandlePanic(err error, args ...interface{}) {
	if err != nil {
		pc, fn, line, _ := runtime.Caller(1)
		logger.Println(fmt.Sprintf("%9v\t%s:%d in %s\t%v", "[PANIC]", cutFilePath(fn), line, cutMethodPath(runtime.FuncForPC(pc).Name()), err.Error()))
		panic(err)
	}
}

func (logger *NewsBotLogger) Info(msg string, args ...interface{}) {
	pc, fn, line, _ := runtime.Caller(1)
	logger.Println(fmt.Sprintf("%9v\t%s:%d in %s\t%v", "[INFO]", cutFilePath(fn), line, cutMethodPath(runtime.FuncForPC(pc).Name()), msg))
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
