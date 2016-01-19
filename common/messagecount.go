package common

type MessageCounter struct {
	InCount  uint64
	OutCount uint64
}

func (m *MessageCounter) In() {
	m.InCount++
}

func (m *MessageCounter) Out() {
	m.OutCount++
}
