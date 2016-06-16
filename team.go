package workers

import (
	"sync"
)

type Team struct {
	mutex sync.RWMutex

	members   []*Worker
	positions map[string]int
}

func NewTeam() *Team {
	return &Team{
		members:   make([]*Worker, 0),
		positions: map[string]int{},
	}
}

func (p *Team) Len() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return len(p.members)
}

func (p *Team) Less(i, j int) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.members[i].status < p.members[j].status
}

func (p *Team) Swap(i, j int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	n := len(p.members)

	if i >= 0 && i < n && j >= 0 && j < n {
		p.members[i], p.members[j] = p.members[j], p.members[i]
	}
}

func (p *Team) Push(x interface{}) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	worker := x.(*Worker)
	p.positions[worker.id] = len(p.members)

	p.members = append(p.members, worker)
}

func (p *Team) Pop() interface{} {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	n := len(p.members)

	member := p.members[n-1]
	p.positions[member.id] = -1

	p.members = p.members[0 : n-1]

	return member
}

func (p *Team) GetMembers() []*Worker {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.members
}

func (p *Team) GetIndexById(id string) int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if v, ok := p.positions[id]; ok {
		return v
	}

	return -1
}
