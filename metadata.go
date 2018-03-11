package workers

type MetadataKey int64

const (
	DispatcherMetadataStatus MetadataKey = iota
)

const (
	WorkerMetadataStatus MetadataKey = iota
	WorkerMetadataTask
	WorkerMetadataLocked
)

const (
	TaskMetadataStatus MetadataKey = iota
	TaskMetadataAttempts
	TaskMetadataAllowStartAt
	TaskMetadataFirstStartedAt
	TaskMetadataLastStartedAt
	TaskMetadataLocked
)

const (
	ListenerMetadataFires MetadataKey = iota
	ListenerMetadataFirstFiredAt
	ListenerMetadataLastFireAt
	ListenerMetadataEvents
)

type Metadata map[MetadataKey]interface{}
