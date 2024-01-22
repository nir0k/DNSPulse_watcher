package syncing

import (
	"HighFrequencyDNSChecker/components/datastore"
	"HighFrequencyDNSChecker/components/logger"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)


func CheckMembersConfig() {
	var (
		membersState []datastore.SyncMemberStateStruct
		newestConfig datastore.StateStruct
	)
	conf := datastore.GetConfig().Sync
	if !conf.SyncEnabled {
		logger.Logger.Debug("Check new configuration skipped. Sycn disabled in configuration")
		return
	}
	members := datastore.GetConfig().SyncMembers
	newestConfig.Configuration = datastore.GetConfHash()
	newestConfig.Csv = datastore.GetCSVHash()
	for _, m := range members {
		var memberState datastore.SyncMemberStateStruct
		url := fmt.Sprintf("https://%s:%d/api/conf/state", m.Host, m.Port)
        memberConfig, err := fetchConfig(url, conf.Token)
        if err !=nil {
			logger.Logger.Error("error fetching data from '", m, "': ", err)
            continue
        }
		memberState.LastCheckDate = time.Now().Unix()
		memberState.Hostname = m.Host
		memberState.Port = m.Port
		memberState.State = memberConfig
		membersState = append(membersState,memberState)
		if newestConfig.Configuration.LastHash != memberConfig.Configuration.LastHash {
			if newestConfig.Configuration.LastUpdate < memberConfig.Configuration.LastUpdate {
				newestConfig.Configuration = memberConfig.Configuration
			}
		}

		if newestConfig.Csv.LastHash != memberConfig.Csv.LastHash {
			if newestConfig.Csv.LastUpdate < memberConfig.Csv.LastUpdate {
				newestConfig.Csv = memberConfig.Csv
			}
		}
	}
	datastore.SetSyncMembersState(membersState)
}

func fetchConfig(url string, token string) (datastore.StateStruct, error) {
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return datastore.StateStruct{}, err
    }

	req.Header.Set("Authorization", token)

	resp, err := client.Do(req)
    if err != nil {
        return datastore.StateStruct{}, err
    }
    defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
        return datastore.StateStruct{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
	logger.Logger.Debug("Raw API Response: ", string(body))
    if err != nil {
        return datastore.StateStruct{}, fmt.Errorf("error reading response body: %w", err)
    }

    var stateConf datastore.StateStruct
    if err := json.Unmarshal(body, &stateConf); err != nil {
        return datastore.StateStruct{}, fmt.Errorf("error unmarshalling JSON: %w, body: %s", err, string(body))
    }

    return stateConf, nil
}