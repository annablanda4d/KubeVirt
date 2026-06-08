package migration

import (
	"errors"
	"time"
)

type DomainJobInfo struct {
	DataRemaining uint64
	DataTotal     uint64
	DataProcessed uint64
}

type LibvirtClient interface {
	GetJobStats() (*DomainJobInfo, error)
	AbortJob() error
}

type MigrationMonitor struct {
	Client          LibvirtClient
	ProgressTimeout time.Duration
}

func NewMigrationMonitor(client LibvirtClient, progressTimeout time.Duration) *MigrationMonitor {
	return &MigrationMonitor{
		Client:          client,
		ProgressTimeout: progressTimeout,
	}
}

func (m *MigrationMonitor) MonitorMigration(stopChan <-chan struct{}) error {
	var lastRemaining uint64
	var lastProgressTime time.Time
	firstRun := true

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return nil
		case <-ticker.C:
			stats, err := m.Client.GetJobStats()
			if err != nil {
				return err
			}

			now := time.Now()
			if firstRun {
				lastRemaining = stats.DataRemaining
				lastProgressTime = now
				firstRun = false
				continue
			}

			if stats.DataRemaining < lastRemaining {
				lastRemaining = stats.DataRemaining
				lastProgressTime = now
			} else {
				if now.Sub(lastProgressTime) > m.ProgressTimeout {
					err := m.Client.AbortJob()
					if err != nil {
						return err
					}
					return errors.New("migration stalled: progress timeout reached")
				}
			}
		}
	}
}
