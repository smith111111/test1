#!/bin/bash

path="/api/push/msgList?msg_types=1,2,3,4,5&msgid=1"

curl -v -H "AppKey: 40b3409b84ce4177b30056605e36e785" -H "Authorization: test1 test2" "dev.galaxyotc.com:8024${path}"