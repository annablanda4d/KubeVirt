package v1

type MigrationConfiguration struct {
	ProgressTimeoutInSeconds   *int64 `json:"progressTimeoutInSeconds,omitempty"`
	CompletionTimeoutInSeconds *int64 `json:"completionTimeoutInSeconds,omitempty"`
}
