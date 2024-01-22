package api

import (
	"HighFrequencyDNSChecker/components/datastore"
	"HighFrequencyDNSChecker/components/logger"
	"HighFrequencyDNSChecker/components/tools"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)


func ConfigDownloadHandler(w http.ResponseWriter, r *http.Request) {
    var err error

    filePath := datastore.GetConfigFilePath()
	
    file, err := os.Open(filePath)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    defer file.Close()

    // Extract filename for Content-Disposition header
    _, fileName := path.Split(filePath)

    w.Header().Set("Content-Type", "application/x-yaml")
    w.Header().Set("Content-Disposition", "attachment; filename="+fileName)

    _, err = io.Copy(w, file)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    logger.Audit.WithFields(logrus.Fields{
        "action": "config_download",
        "file":   fileName,
        "ip":     tools.GetClientIP(r),
        "token":  tools.ExtractToken(r),
    }).Info("Configuration file downloaded")
}
