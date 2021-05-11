package cache

import (
	"fmt"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestNewCache(t *testing.T) {
	c := NewCache(int64(100), nil, nil)
	c.Set("hi", String("zach"))
	v, ok := c.Get("hi")
	if ok {
		fmt.Println(v)
	}
}
