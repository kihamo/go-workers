package workers

type Manager interface {
	Push(ManagerItem) error
	Pull() ManagerItem
	Remove(ManagerItem)
	GetAll() []ManagerItem
	Check() bool
}
