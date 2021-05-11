package zcache

// pick node
type PeerPicker interface {
	Pick(key string) (PeerGetter, bool)
}

// get data from specific group
type PeerGetter interface {
	Get(group, key string) ([]byte, error)
}
