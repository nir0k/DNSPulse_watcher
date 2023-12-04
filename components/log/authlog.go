package log

import (
    "github.com/sirupsen/logrus"
    "os"
)

type AuthLogHook struct {
    Writer *os.File
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
        _, err = hook.Writer.Write([]byte(line))
        return err
    }
    return nil
}


func InitAuthLog(filepath string, level logrus.Level) (*logrus.Logger) {
    authLog := logrus.New()

    file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
    if err != nil {
        logrus.Fatalf("Error opening file '%s': %v", filepath, err)
    }

    authLog.SetFormatter(&logrus.JSONFormatter{})
    authLog.SetOutput(file)
    authLog.SetLevel(level)

    AuthLog = authLog

	return authLog
}