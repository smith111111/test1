#!/bin/bash

path="/rpc/push/sendMsg"

body='{
	"userids": "10001",
	"appid":3,
	"title":"收到了吗？",
	"text":"收到请告诉我！",
	"custom":{"msg_type": 1, "msg_body": {"orderid”:10001, "status": 1}},
	"display_type":"message"
}'

curl -v -H "AppKey: 40b3409b84ce4177b30056605e36e785" -H "Authorization: test1 test2" -k -d "${body}" "dev.galaxyotc.com:8024${path}"