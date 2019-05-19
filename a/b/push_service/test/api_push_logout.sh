#!/bin/bash

path="/rpc/push/logout"

body='{
	"userid": 10001,
	"appid":1
}'

curl -v -H "AppKey: 40b3409b84ce4177b30056605e36e785" -H "Authorization: test1 test2" -k -d "${body}" "dev.galaxyotc.com:8024${path}"