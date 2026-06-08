package migration

import (
	"testing"
	"time"
)

type MockTestLibvirtClient struct {
	dataRemaining uint64
	aborted       bool
}

func (m *MockTestLibvirtClient) GetJobStats() (*DomainJobInfo, error) {
	return &DomainJobInfo{
		DataRemaining: m.dataRemaining,
		DataTotal:     100,
		DataProcessed: 100 - m.dataRemaining,
	}, nil
}

func (m *MockTestLibvirtClient) AbortJob() error {
	m.aborted = true
	return nil
}

func TestMigrationMonitorStall(t *testing.T) {
	client := &MockTestLibvirtClient{
		dataRemaining: 50,
	}
	monitor := NewMigrationMonitor(client, 100*time.Millisecond)
	stopChan := make(chan struct{})

	err := monitor.MonitorMigration(stopChan)
	if err == nil {
		t.Fatal("expected error due to migration stall, got nil")
	}

	if err.Error() != "migration stalled: progress timeout reached" {
		t.Errorf("expected stall error, got: %v", err)
	}

	if !client.aborted {
		t.Error("expected libvirt job to be aborted")
	}
}
