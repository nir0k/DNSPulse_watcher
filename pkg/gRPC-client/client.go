package grpcclient

import (
	"DNSPulse_watcher/pkg/datastore"
	"DNSPulse_watcher/pkg/logger"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	pb "DNSPulse_watcher/pkg/gRPC"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)


func FetchConfig(server datastore.ConfigHubConfigStruct) (bool, bool, error) {
	var (
		segmentConfStatus bool
		pollingHostsStatus bool
		err error
	)
	srvURL := fmt.Sprintf("%s:%d", server.Host, server.Port)
	conn, err := grpc.Dial(srvURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Logger.Errorf("did not connect: %v", err)
		return segmentConfStatus, pollingHostsStatus, err
	}
	defer conn.Close()
	c := pb.NewConfigHubServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	md := metadata.New(map[string]string{"authorization": server.Token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	r1, err := c.GetSegmentConfig(ctx, &pb.GetSegmentConfigRequest{Token: server.Token, SegmentName: server.Segment})
	if err != nil {
		logger.Logger.Errorf("could not fetch segment config: %v", err)
		segmentConfStatus = false
	}
	segmentConfStatus = true
	logger.Logger.Printf("Segment Config: %s", r1)
	path := filepath.Join(server.Path, "segmentConfig.json")
	if err := writeJSONToFile(path, r1); err != nil {
        logger.Logger.Errorf("Failed to write configuration to JSON file: %v", err)
    }
	datastore.SetSegmentConfig(r1)

	r2, err := c.GetCsv(ctx, &pb.GetCsvRequest{Token: server.Token, SegmentName: server.Segment})
	if err != nil {
		logger.Logger.Errorf("could not fetch CSV: %v", err)
		pollingHostsStatus = false
	}
	pollingHostsStatus = true
	path = filepath.Join(server.Path, "pollingHosts.json")
	if err := writeJSONToFile(path, r2); err != nil {
        logger.Logger.Errorf("Failed to write polling hosts to JSON file: %v", err)
    }

	_, err = datastore.LoadPollingHosts(r2.Csvs)
	if err != nil {
		logger.Logger.Errorf("Failed to load polling hosts: %v", err)
	}
	return segmentConfStatus, pollingHostsStatus, err

}

func writeJSONToFile(filename string, data proto.Message) error {
    marshaler := protojson.MarshalOptions{
        Multiline: true,
        Indent:    "  ",
    }
    out, err := marshaler.Marshal(data)
    if err != nil {
        return err
    }

    if err := os.WriteFile(filename, out, 0644); err != nil {
        return err
    }
    return nil
}
