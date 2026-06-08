package main

import (
	"fmt"
	"time"

	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/pkg/virt-controller/watch"
	"kubevirt.io/pkg/virt-handler/migration"
)

type MockLibvirtClient struct {
	dataRemaining uint64
	aborted       bool
	stallAfter    int
	ticks         int
}

func (m *MockLibvirtClient) GetJobStats() (*migration.DomainJobInfo, error) {
	m.ticks++
	if m.ticks <= m.stallAfter {
		if m.dataRemaining > 10 {
			m.dataRemaining -= 10
		}
	}
	return &migration.DomainJobInfo{
		DataRemaining: m.dataRemaining,
		DataTotal:     100,
		DataProcessed: 100 - m.dataRemaining,
	}, nil
}

func (m *MockLibvirtClient) AbortJob() error {
	m.aborted = true
	return nil
}

func main() { 
	fmt.Println("Starting KubeVirt Live Migration Stall Detection Simulation...")

	timeoutVal := int64(1)
	config := &v1.MigrationConfiguration{
		ProgressTimeoutInSeconds: &timeoutVal,
	}

	progressTimeout := time.Duration(*config.ProgressTimeoutInSeconds) * time.Second

	client := &MockLibvirtClient{
		dataRemaining: 100,
		stallAfter:    2,
	}

	monitor := migration.NewMigrationMonitor(client, progressTimeout)
	stopChan := make(chan struct{})

	fmt.Println("Monitoring migration...")
	err := monitor.MonitorMigration(stopChan)
	if err != nil {
		fmt.Printf("Migration monitoring stopped with error: %v\n", err)
	}

	vmim := &watch.VirtualMachineInstanceMigration{}
	controller := &watch.MigrationController{}
	controller.ReconcileMigration(vmim, err)

	fmt.Printf("VMIM Status Phase: %s\n", vmim.Status.Phase)
	fmt.Printf("VMIM Status Reason: %s\n", vmim.Status.Reason)
	fmt.Printf("VMIM Status Message: %s\n", vmim.Status.Message)
	fmt.Printf("Libvirt Job Aborted: %t\n", client.aborted)

	if vmim.Status.Phase == watch.MigrationFailed &&
		vmim.Status.Reason == "MigrationStalledDueToNetworkCongestion" &&
		client.aborted {
		fmt.Println("Simulation SUCCESS: Migration stall detected and handled correctly!")
	} else {
		fmt.Println("Simulation FAILED: Migration stall not handled correctly.")
	}
}
