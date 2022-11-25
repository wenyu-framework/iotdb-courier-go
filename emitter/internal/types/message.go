package types

type Message struct {
	Partition int32
	Offset    int64
	Key       string
	Tag       string
	Value     []byte
}
