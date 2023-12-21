package watcher

import (
	"context"
	"io"
	"reflect"
	"testing"
	"time"

	sqldb "HighFrequencyDNSChecker/components/db"
	log "HighFrequencyDNSChecker/components/log"

	"github.com/castai/promwrite"
	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
)

// Mock data for PrometheusLabelConfiguration
var mockPrometheusLabel = sqldb.PrometheusLabelConfiguration{
    // Populate this with the relevant mock values
    // Example: Truncated: true, Opcode: false, etc.
}

type MockClient struct {
    WriteFunc func(ctx context.Context, req *promwrite.WriteRequest, headers map[string]string) (int, error)
}

func (m *MockClient) Write(ctx context.Context, req *promwrite.WriteRequest, headers map[string]string) (int, error) {
    return m.WriteFunc(ctx, req, headers)
}


// Example test for collectLabels function
func TestCollectLabels(t *testing.T) {
    // Setup test resolver
    testResolver := sqldb.Resolver{
        Server: "testServer",
        IPAddress: "192.168.1.1",
        Domain: "example.com",
        Location: "testLocation",
        Site: "testSite",
        Protocol: "UDP",
        ServerSecurityZone: "testZone",
        Zonename: "testZoneName",
        ServiceMode: true,
		Recursion: true,
        // Other fields as necessary
    }

	mockPrometheusLabel = sqldb.PrometheusLabelConfiguration{
        Opcode:             false,
		Authoritative:      false,
		Truncated:          true,
		Rcode:              true,
		RecursionDesired:   false,
		RecursionAvailable: false,
		AuthenticatedData:  false,
		CheckingDisabled:   false,
		PollingRate:        false,
		Recursion:          true,
    }

    // Setup test DNS message header
    testMsgHdr := dns.MsgHdr{
		Truncated: false,
        Rcode:     10,
    }

    // Assign mock PrometheusLabel for testing
    PrometheusLabel = mockPrometheusLabel

	PrometheusConfig = sqldb.PrometheusConfiguration{
        MetricName: "testMetricName",
        // ... [other fields as necessary] ...
    }

    // Call collectLabels
    gotLabels := collectLabels(testResolver, testMsgHdr, PrometheusLabel)

    // Define expected labels
    expectedLabels := []promwrite.Label{
        {Name: "__name__", Value: "testMetricName"},
        {Name: "server", Value: "testServer"},
        {Name: "server_ip", Value: "192.168.1.1"},
        {Name: "domain", Value: "example.com"},
        {Name: "location", Value: "testLocation"},
        {Name: "site", Value: "testSite"},
        {Name: "watcher", Value: ""},
        {Name: "watcher_ip", Value: ""},
        {Name: "watcher_security_zone", Value: ""},
        {Name: "watcher_location", Value: ""},
        {Name: "protocol", Value: "UDP"},
        {Name: "server_security_zone", Value: "testZone"},
        {Name: "service_mode", Value: "true"},
        {Name: "zonename", Value: "testZoneName"},
        {Name: "rcode", Value: "10"},
		{Name: "truncated", Value: "false"},
		{Name: "recursion", Value: "true"},
        // Add other labels based on the logic in your `collectLabels` function
    }

    // Compare the results
    if !reflect.DeepEqual(gotLabels, expectedLabels) {
        t.Errorf("collectLabels() = %v, want %v", gotLabels, expectedLabels)
    }
}

// TestBufferTimeSeries tests the bufferTimeSeries function.
func TestBufferTimeSeries(t *testing.T) {
    // Setup
    server := sqldb.Resolver{
        // Initialize necessary fields
    }
    header := dns.MsgHdr{
        // Initialize necessary fields
    }
    tm := time.Now()
    value := 1.0 // Example value

    // Reset global Buffer before test
    Buffer = []promwrite.TimeSeries{}

    // Set the buffer size limit for the test
    PrometheusConfig.BuferSize = 3

    // Call bufferTimeSeries three times
    for i := 0; i < 3; i++ {
        BufferTimeSeries(server, tm, value, header)
    }

    // Assertions
    if len(Buffer) != 3 {
        t.Errorf("Buffer was not updated correctly, got size = %d, want %d", len(Buffer), 3)
    }

    // Test that Buffer is reset and sendVM is called after fourth invocation
    BufferTimeSeries(server, tm, value, header)
    if len(Buffer) != 0 {
        t.Errorf("Buffer was not reset correctly after reaching limit, got size = %d, want %d", len(Buffer), 0)
    }

    // Additional assertions as needed...
}

func TestSendVM(t *testing.T) {
	log.AppLog = logrus.New()
    log.AppLog.Out = io.Discard
    // Backup the original PrometheusConfig
    originalConfig := PrometheusConfig

    // Mock setup
    PrometheusConfig = sqldb.PrometheusConfiguration{
        Url: "http://mockserver.com",
        RetriesCount: 3,
    }

    // Prepare mock time series data
    timeSeries := make([]promwrite.TimeSeries, 1)
    timeSeries[0] = promwrite.TimeSeries{
        Labels: []promwrite.Label{
            {Name: "test_label", Value: "test_value"},
        },
        Sample: promwrite.Sample{
            Time:  time.Now(),
            Value: 1.23,
        },
    }

    // Execute sendVM
    result := sendVM(timeSeries)

    // Assert that sendVM returns a boolean
    if _, ok := interface{}(result).(bool); !ok {
        t.Errorf("sendVM should return a boolean, got %T", result)
    }

    // Additional assertions can be made here if possible

    // Restore the original PrometheusConfig
    PrometheusConfig = originalConfig
}