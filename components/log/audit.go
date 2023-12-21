package log

import (
	sqldb "HighFrequencyDNSChecker/components/db"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var AuthLog *logrus.Logger

type AuthLogHook struct {
    Logger *lumberjack.Logger
}

func (hook *AuthLogHook) Levels() []logrus.Level {
    return logrus.AllLevels
}

func (hook *AuthLogHook) Fire(entry *logrus.Entry) error {
    if _, ok := entry.Data["username"]; ok {
        line, err := entry.String()
        if err != nil {
            return err
        }
        _, err = hook.Logger.Write([]byte(line))
        return err
    }
    return nil
}


func InitAuthLog() (*logrus.Logger, error) {
	conf, err := sqldb.GetLogConfig(sqldb.AppDB, 1)
    if err != nil {
        return nil, err
    }
    lumberjackLogger := &lumberjack.Logger{
        Filename:   conf.Path,
        MaxSize:    conf.MaxSize,
        MaxBackups: conf.MaxFiles,
        MaxAge:     conf.MaxAge,
        Compress:   true,
    }
	severity, _ := GetLogSeverity(conf.MinSeverity)
    authLog := logrus.New()
    authLog.SetOutput(lumberjackLogger)
    authLog.SetFormatter(&logrus.JSONFormatter{})
    authLog.SetLevel(severity)

    hook := &AuthLogHook{Logger: lumberjackLogger}
    authLog.AddHook(hook)

    AuthLog = authLog

    return authLog, nil
}