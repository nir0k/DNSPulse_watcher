package log

import (
    "github.com/sirupsen/logrus"
    "os"
)

func InitAppLogger(filepath string, level logrus.Level) *logrus.Logger {
    logger := logrus.New()

    file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
    if err != nil {
        logrus.Fatalf("Error opening file '%s': %v", filepath, err)
    }

    logger.SetFormatter(&logrus.JSONFormatter{})
    logger.SetOutput(file)
    logger.SetLevel(level)

	AppLog = logger

    return logger
}
