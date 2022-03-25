package kafkatest

type MessageRecord struct {
	Topic string
	Payload interface{}
}

type MessageRecorder interface {
	Reset()
	Records(topic string) []*MessageRecord
	AllRecords() []*MessageRecord
}

type messageRecorder interface {
	MessageRecorder
	Record(msg *MessageRecord)
}
