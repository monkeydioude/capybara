package grpc

type ProxyMessage struct {
	data []byte
}

// Implement the proto.Message interface
func (m *ProxyMessage) Reset()         { *m = ProxyMessage{} }
func (m *ProxyMessage) String() string { return "DummyMessage" }
func (m *ProxyMessage) ProtoMessage()  {}
func (m *ProxyMessage) Marshal() ([]byte, error) {
	return m.data, nil
}
func (m *ProxyMessage) Unmarshal(data []byte) error {
	m.data = data
	return nil
}
