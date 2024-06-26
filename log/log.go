package log

import (
	"io"
	"log"
	"os"
	"sync"
)

var (
	// log.Lshortfile 支持显示文件名和代码行号
	errorLog = log.New(os.Stdin, "\033[31m[error]\033[0m", log.LstdFlags|log.Llongfile)
	infoLog  = log.New(os.Stdin, "\033[34m[info ]\033[0m", log.LstdFlags|log.Llongfile)
	loggers  = []*log.Logger{errorLog, infoLog}
	mu       sync.Mutex
)

var (
	Error  = errorLog.Println
	ErrorF = errorLog.Printf
	Info   = infoLog.Println
	InfoF  = infoLog.Printf
)

// log levels
const (
	InfoLevel = iota
	ErrorLevel
	Disables
)

func SetLevel(level int) {
	mu.Lock()
	defer mu.Unlock()
	// 三个层级声明为三个常量，通过控制 Output，来控制日志是否打印
	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}
	//如果设置为 ErrorLevel，infoLog 的输出会被定向到 ioutil.Discard，即不打印该日志
	if ErrorLevel < level {
		errorLog.SetOutput(io.Discard)
	}
	if InfoLevel < level {
		infoLog.SetOutput(io.Discard)
	}
}
