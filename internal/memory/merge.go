package memory

type Merger struct{}

type MemoryItem struct {
	ID      string
	Content string
}

func NewMerger() *Merger {
	return &Merger{}
}

func (m *Merger) Merge(mem1, mem2 *MemoryItem) *MemoryItem {
	if mem1 == nil {
		return mem2
	}
	if mem2 == nil {
		return mem1
	}

	if mem1.ID == mem2.ID {
		return mem1
	}

	return &MemoryItem{
		ID:      mem1.ID + "_" + mem2.ID,
		Content: mem1.Content + " | " + mem2.Content,
	}
}

func (m *Merger) ResolveConflict(mem1, mem2 *MemoryItem) *MemoryItem {
	if mem1 == nil || mem2 == nil {
		return mem1
	}

	if mem1.ID != mem2.ID {
		return mem1
	}

	if len(mem1.Content) >= len(mem2.Content) {
		return mem1
	}
	return mem2
}