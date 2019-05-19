#!/bin/bash

path="/api/push/deviceInfo"

body='{
	"appid":3,
	"device_token":"12345-12345-12345-12345",
	"push_type":1
}'

curl -v -H "AppKey: 40b3409b84ce4177b30056605e36e785" -H "Authorization: test1 test2" -k -d "${body}" "dev.galaxyotc.com:8024${path}"