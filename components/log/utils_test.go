package log

import (
    "testing"

    "github.com/sirupsen/logrus"
)

func TestGetLogSeverity(t *testing.T) {
    testCases := []struct {
        name     string
        severity string
        expected logrus.Level
        isValid  bool
    }{
        {"Debug Level", "debug", logrus.DebugLevel, true},
        {"Info Level", "info", logrus.InfoLevel, true},
        {"Warning Level", "warning", logrus.WarnLevel, true},
        {"Error Level", "error", logrus.ErrorLevel, true},
        {"Fatal Level", "fatal", logrus.FatalLevel, true},
        {"Invalid Level", "invalid", logrus.WarnLevel, false},
        // Add more test cases if necessary
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            level, valid := GetLogSeverity(tc.severity)
            if level != tc.expected || valid != tc.isValid {
                t.Errorf("GetLogSeverity(%s) = %v, %v; want %v, %v", tc.severity, level, valid, tc.expected, tc.isValid)
            }
        })
    }
}
