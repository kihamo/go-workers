package workers

type MetadataKey int64

const (
	DispatcherMetadataWorkersWaiting MetadataKey = iota
	DispatcherMetadataTasksWaiting
)

const (
	WorkerMetadataStatus MetadataKey = iota
	WorkerMetadataTask
)

const (
	TaskMetadataStatus MetadataKey = iota
	TaskMetadataAttempts
	TaskMetadataAllowStartAt
	TaskMetadataStartedAt
	TaskMetadataFinishedAt
)

type Metadata map[MetadataKey]interface{}