package log

import (
	sqldb "HighFrequencyDNSChecker/components/db"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var AppLog *logrus.Logger

func InitAppLogger() (*logrus.Logger, error) {
    conf, err := sqldb.GetLogConfig(sqldb.AppDB, 0)
    if err != nil {
        return nil, err
    }
    logger := logrus.New()
    logger.SetOutput(&lumberjack.Logger{
        Filename:   conf.Path,
        MaxSize:    conf.MaxSize,
        MaxBackups: conf.MaxFiles,
        MaxAge:     conf.MaxAge,
        Compress:   true,
    })

    severity, _ := GetLogSeverity(conf.MinSeverity)

    logger.SetFormatter(&logrus.JSONFormatter{})
    logger.SetLevel(severity)

	AppLog = logger
    return logger, nil
}
