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

func init() {
	fileRotation := &lumberjack.Logger{
		Filename:   consts.ProjectName + ".log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     3,    //days
		Compress:   true, // disabled by default
	}
	mw := io.MultiWriter(os.Stderr, fileRotation)
	log.SetOutput(mw)
}

func HandleError(err error) (wasError bool) {
	if err != nil {
		// notice that we're using 1, so it will actually log the where
		// the error happened, 0 = this function, we don't want that.
		pc, fn, line, _ := runtime.Caller(1)
		log.Println(fmt.Sprintf("[ERROR]\t%s:%d in %s\t%v", cutFilePath(fn), line, runtime.FuncForPC(pc).Name(), err.Error()))
		wasError = true
	}
	return wasError
}

func HandlePanic(err error) {
	if err != nil {
		pc, fn, line, _ := runtime.Caller(1)
		log.Println(fmt.Sprintf("[PANIC]\t%s:%d in %s\t%v", cutFilePath(fn), line, runtime.FuncForPC(pc).Name(), err.Error()))
		panic(err)
	}
}

func Info(msg string) {
	pc, fn, line, _ := runtime.Caller(1)
	log.Println(fmt.Sprintf("[INFO]\t%s:%d in %s\t%v", cutFilePath(fn), line, runtime.FuncForPC(pc).Name(), msg))
}

func cutFilePath(fn string) string {
	i := strings.Index(fn, consts.ProjectName)
	if i > 0 {
		return fn[i:]
	}
	return fn
}
