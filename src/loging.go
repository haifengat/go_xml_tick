package src

import (
	"time"

	rotates "github.com/lestrrat/go-file-rotatelogs"
	"github.com/pkg/errors"
	"github.com/rifflock/lfshook"
	logger "github.com/sirupsen/logrus"
)

// LogInit 日志初始化
func LogInit() {
	logger.SetLevel(logger.InfoLevel)
	logger.AddHook(newRotateHook("", "stdout.logger", 7*24*time.Hour, 24*time.Hour))
}

func newRotateHook(logPath string, logFileName string, maxAge time.Duration, rotationTime time.Duration) *lfshook.LfsHook {
	baseLogPath := logPath //path.Join(logPath, logFileName)

	writer, err := rotates.New(
		baseLogPath+"%Y-%m-%d.log",
		rotates.WithLinkName(baseLogPath),      // 生成软链，指向最新日志文
		rotates.WithMaxAge(maxAge),             // 文件最大保存时间
		rotates.WithRotationTime(rotationTime), // 日志切割时间间隔
	)
	if err != nil {
		logger.Errorf("config local file system logger error. %+v", errors.WithStack(err))
	}
	return lfshook.NewHook(lfshook.WriterMap{
		logger.DebugLevel: writer, // 为不同级别设置不同的输出目的
		logger.InfoLevel:  writer,
		logger.WarnLevel:  writer,
		logger.ErrorLevel: writer,
		logger.FatalLevel: writer,
		logger.PanicLevel: writer,
	}, &logger.TextFormatter{DisableColors: true, TimestampFormat: "2006-01-02 15:04:05"})
}
