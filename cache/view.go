package cache

type View struct {
	data []byte
}

func (v View) Len() int {
	return len(v.data)
}

func (v View) String() string {
	return string(v.data)
}

func (v View) ByteSlice() []byte {
	return cloneBytes(v.data)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
