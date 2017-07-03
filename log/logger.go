package log

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
)

func init() {
	LogInit("/apps/logs/husky/stdout.log")
}

//统一日志接口
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Print(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Println(args ...interface{})
	Warnln(args ...interface{})
	Warningln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
	Panicln(args ...interface{})
}

//错误信息收集器
var errorCollectors = make([]string, 0, 1000)

//获取日志输出器
func GetLogger() Logger {
	return defaultLogger
}

func GetConsoleLogger() Logger {
	return consoleLogger
}

//自定义的默认钩子
type DefaultFieldsHook struct {
}

//log's on fire
func (df *DefaultFieldsHook) Fire(entry *logrus.Entry) error {
	//如果是错误的日志，则寄存一份在ErrorCollectors中
	display := fmt.Sprintf("[%s][ERROR] %s",
		entry.Time.Format("2006-01-02 15:04:05"), entry.Message)
	errorCollectors = append(errorCollectors, display)
	return nil
}

func (df *DefaultFieldsHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.ErrorLevel,
	}
}

//默认日志输出
var defaultLogger *logrus.Logger
var consoleLogger *logrus.Logger

func LogInit(logpath string) {
	if logpath == "" {
		logpath = "./stdout.log"
	}
	defaultLogger = logrus.New()
	consoleLogger = logrus.New()
	SetLogOut(defaultLogger, logpath)
	SetLogOut(consoleLogger, "stdout")
	SetLogLevel(defaultLogger, "debug")
	SetLogLevel(consoleLogger, "debug")
	defaultLogger.Formatter = &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}

	//日志钩子
	//defaultLogger.Hooks.Add(&DefaultFieldsHook{})

	consoleLogger.Formatter = &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}
}

//设置日志输出方式
func SetLogOut(log *logrus.Logger, outString string) error {
	switch outString {
	case "stdout":
		log.Out = os.Stdout
	case "stderr":
		log.Out = os.Stderr
	default:
		if checkLogFileExist(outString) != nil {
			filepath := filepath.Dir(outString)
			fmt.Printf("create log file in %s\n", filepath)
			cmd := "mkdir -p " + filepath
			exec.Command("sh", "-c", cmd).Run()
		}
		f, err := os.OpenFile(outString, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)

		if err != nil {
			println(err.Error())
			return err
		}

		log.Out = f
	}

	return nil
}

//设置日志级别
func SetLogLevel(log *logrus.Logger, levelString string) error {
	level, err := logrus.ParseLevel(levelString)

	if err != nil {
		return err
	}

	log.Level = level
	return nil
}

func ErrorLogsSize() int {
	if errorCollectors == nil {
		return 0
	}
	return len(errorCollectors)
}

//错误记录一次性输出
func ShuffleErrors() {
	for _, msg := range errorCollectors {
		fmt.Println(msg)
	}
	errorCollectors = errorCollectors[:0]
}

func checkLogFileExist(filePath string) error {

	if filePath == "" {
		return errors.New("数据文件路径为空")
	}

	if _, err := os.Stat(filePath); err != nil {
		return errors.New("PathError:" + err.Error())
	}
	return nil
}
