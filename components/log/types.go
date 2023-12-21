package log

import "github.com/sirupsen/logrus"

type Log_conf struct {
    Path string
    Level logrus.Level
    MaxAge int
    MaxSize int
    MaxFiles int
}