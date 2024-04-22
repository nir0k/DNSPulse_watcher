package datastore

import (
	pb "DNSPulse_watcher/pkg/gRPC"
	"DNSPulse_watcher/pkg/logger"
	"sync"
)


var (
	segmentConfig *pb.SegmentConfStruct
	segmentConfigMutex sync.RWMutex
)

func GetSegmentConfig() *pb.SegmentConfStruct {
	segmentConfigMutex.RLock()
    defer segmentConfigMutex.RUnlock()
	return segmentConfig
}

func SetSegmentPollingHash(hash string) {
	segmentConfigMutex.Lock()
    defer segmentConfigMutex.Unlock()
	segmentConfig.Polling.Hash = hash
}

func SetSegmentConfig(newConf *pb.SegmentConfStruct) {
	segmentConfigMutex.Lock()
    defer segmentConfigMutex.Unlock()
	if segmentConfig == nil || segmentConfig.General == nil {
		segmentConfig = newConf
		logger.Logger.Infof("Loaded new segment config")
		return
	}
	if segmentConfig.General.Hash != newConf.General.Hash {
		segmentConfig = newConf
		logger.Logger.Infof("Loaded and used new segment config")
	}
	logger.Logger.Infof("Segment config not changes")
}
