package log

import (
    "github.com/sirupsen/logrus"
    "gopkg.in/natefinch/lumberjack.v2"
    // "os"
)

func InitAppLogger(conf Log_conf, level logrus.Level) *logrus.Logger {
    logger := logrus.New()
    logger.SetOutput(&lumberjack.Logger{
        Filename:   conf.Log_path,
        MaxSize:    conf.Log_max_size,
        MaxBackups: conf.Log_max_files,
        MaxAge:     conf.Log_max_age,
        Compress:   true,
    })

    // file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
    // if err != nil {
    //     logrus.Fatalf("Error opening file '%s': %v", filepath, err)
    // }

    logger.SetFormatter(&logrus.JSONFormatter{})
    // logger.SetOutput(file)
    logger.SetLevel(level)

	AppLog = logger

    return logger
}
