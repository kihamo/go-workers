package workers

type MetadataKey int64

const (
	DispatcherMetadataWorkersUnlocked MetadataKey = iota
	DispatcherMetadataTasksUnlocked
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

const (
	ListenerMetadataFires MetadataKey = iota
	ListenerMetadataFirstFiredAt
	ListenerMetadataLastFireAt
	ListenerMetadataEventIds
)

type Metadata map[MetadataKey]interface{}
