package node

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"zcache/zlog"
)

type ClientNode struct {
	BaseURL string
}

func (c *ClientNode) Get(group, key string) ([]byte, error) {
	// 向server发起请求
	u := fmt.Sprintf(
		"%v/%v/get/%v",
		c.BaseURL,
		group,
		key,
	)
	zlog.Info("request url:" + u)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}
