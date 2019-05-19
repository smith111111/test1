package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

//post访问方式(推荐)
//host 请求地址
//param 附加参数
func RequestPost(host string, param string) (string, error){
	body := ioutil.NopCloser(strings.NewReader(param))
	req, err := http.NewRequest("POST", host, body)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(0)
	}
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Encoding", "UTF-8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8") //这个一定要加，不加form的值post不过去
	//fmt.Printf("%+v\n", req)      //看下发送的结构

	client := &http.Client{}
	resp, err := client.Do(req) //发送
	if err != nil {
		fmt.Println("error:", err)
		return "",err
	}
	defer resp.Body.Close() //一定要关闭resp.Body
	data, err := ioutil.ReadAll(resp.Body)
	return string(data), err
}

//get方式请求
//host 请求地址
//param 附加参数
func RequestGet(host string, param string) (string){
	req, err := http.NewRequest("GET", host+"?"+param, nil)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(0)
	}
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Encoding", "UTF-8")
	req.Header.Add("Authorization", "APPCODE " + "c0ae540d02044b43a12aba5d732b1e0a")
	//fmt.Printf("%+v\n", req)      //看下发送的结构

	client := &http.Client{}
	resp, err := client.Do(req) //发送
	if err != nil {
		fmt.Println("error:", err)
		return ""
	}
	defer resp.Body.Close() //一定要关闭resp.Body
	data, err := ioutil.ReadAll(resp.Body)

	return string(data)
}