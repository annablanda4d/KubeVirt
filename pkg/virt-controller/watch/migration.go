package watch

type MigrationState string

const (
	MigrationFailed MigrationState = "Failed"
)

type VirtualMachineInstanceMigrationStatus struct {
	Phase   MigrationState `json:"phase,omitempty"`
	Reason  string         `json:"reason,omitempty"`
	Message string         `json:"message,omitempty"`
}

type VirtualMachineInstanceMigration struct {
	Status VirtualMachineInstanceMigrationStatus `json:"status,omitempty"`
}

type MigrationController struct{}

func (c *MigrationController) ReconcileMigration(vmim *VirtualMachineInstanceMigration, err error) {
	if err != nil && err.Error() == "migration stalled: progress timeout reached" {
		vmim.Status.Phase = MigrationFailed
		vmim.Status.Reason = "MigrationStalledDueToNetworkCongestion"
		vmim.Status.Message = "Virtual machine migration stalled due to network congestion or packet loss."
	}
}
