package workers

type Status interface {
	Int64() int64
	String() string
}
