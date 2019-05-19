package search_service

import (
	"testing"
)

func newClient() *Client {
	addrs := []string{"http://192.168.0.231:9200"}

	return NewClient(addrs)
}

func TestXiaoMiPush(t *testing.T) {
	c := newClient()

	userId := uint64(1)

	userInfo1 := &SearchUserInfo{
		Id: 1,
		Name: "abc123",
		Mobile: "13025250051",
	}

	err := c.AddUserInfo(userId, userInfo1)
	if err != nil {
		t.Fatal(err)
	}

	data := make(map[string]interface{})
	data["mobile"] = "13333553333"
	err = c.UpdateUserInfo(userId, data)
	if err != nil {
		t.Fatal(err)
	}

	list, err := c.SearchUserInfo("abc", 0, 10)
	if err == nil {
		for _, item := range list {
			t.Logf("result is : %+v", item)
		}
	}
}
