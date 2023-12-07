package log

import (
    "github.com/sirupsen/logrus"
    "gopkg.in/natefinch/lumberjack.v2"
)

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


func InitAuthLog(conf Log_conf, level logrus.Level) (*logrus.Logger) {
    lumberjackLogger := &lumberjack.Logger{
        Filename:   conf.Log_path,
        MaxSize:    conf.Log_max_size,
        MaxBackups: conf.Log_max_files,
        MaxAge:     conf.Log_max_age,
        Compress:   true,
    }
    authLog := logrus.New()
    authLog.SetOutput(lumberjackLogger)
    authLog.SetFormatter(&logrus.JSONFormatter{})
    authLog.SetLevel(level)

    hook := &AuthLogHook{Logger: lumberjackLogger}
    authLog.AddHook(hook)

    return authLog
}