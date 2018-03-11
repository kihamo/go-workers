package workers

type Manager interface {
	Push(ManagerItem) error
	Pull() ManagerItem
	Remove(ManagerItem)
	GetById(string) ManagerItem
	GetAll() []ManagerItem
}
