// @author      Liu Yongshuai<liuyongshuai@hotmail.com>
// @date        2018-10-30 11:14

package goUtils

import (
	"context"
	"fmt"
	"testing"
)

var (
	testUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36"
)

//纯GET请求
func TestHttpClient_Get(t *testing.T) {
	url := "http:/10.96.114.84/add.php?a=1&b=3"
	client := NewHttpClient(url, context.Background())
	client.AddCookie("c1", "v1")
	client.SetReferer("http://baidu.com")
	client.AddHeader("myHeaderKey", "myHeaderValue")
	client.SetHost("test.wendao.com")
	client.SetKeepAlive(true)
	resp, _ := client.Get()
	fmt.Println(string(resp.GetBody()))
	resp, _ = client.Get()
	fmt.Println(string(resp.GetBody()))
	resp, _ = client.Post()
	fmt.Println(string(resp.GetBody()))
}

//POST几个字段信息
func TestHttpClient_Post(t *testing.T) {
	url := "http://10.96.114.84/add.php"
	client := NewHttpClient(url, context.Background())
	client.AddField("a", "s")
	client.AddField("d", "h")
	client.AddCookie("c1", "v1")
	client.SetUserAgent(testUserAgent)
	client.SetReferer("http://baidu.com")
	client.AddHeader("myHeaderKey", "myHeaderValue")
	client.SetHost("test.wendao.com")
	resp, _ := client.Post()
	fmt.Println(string(resp.GetBody()))
	resp, _ = client.Post()
	fmt.Println(string(resp.GetBody()))
	client.SetUrl("http://10.96.114.84/add.php")
	client.SetHost("phpmyadmin.wendao.com")
	resp, _ = client.Post()
	fmt.Println(string(resp.GetBody()))
}

//POST几个字段信息并上传文件
func TestHttpClient_PostUploadFiles(t *testing.T) {
	url := "http://10.96.114.84/add.php"
	client := NewHttpClient(url, context.Background())
	client.AddField("a", "s")
	client.AddField("d", "h")
	client.AddCookie("c1", "v1")
	client.AddFile("abc", "./http_test.go", "my.cnf")
	client.SetUserAgent(testUserAgent)
	client.SetReferer("http://baidu.com")
	client.SetHost("test.wendao.com")
	client.AddHeader("myHeaderKey", "myHeaderValue")
	resp, _ := client.Post()
	fmt.Println(string(resp.GetBody()))
	resp, _ = client.Post()
	fmt.Println(string(resp.GetBody()))
	client.SetUrl("http://10.96.114.84/add.php")
	client.SetHost("phpmyadmin.wendao.com")
	resp, _ = client.Post()
	fmt.Println(string(resp.GetBody()))
}

//纯上传文件
func TestHttpClient_UploadFiles(t *testing.T) {
	url := "http://10.96.114.84/add.php"
	client := NewHttpClient(url, context.Background())
	client.AddCookie("c1", "v1")
	client.AddFile("abc", "./http_client_test.go", "my.cnf")
	client.SetUserAgent(testUserAgent)
	client.SetReferer("http://baidu.com")
	client.AddHeader("myHeaderKey", "myHeaderValue")
	client.SetHost("test.wendao.com")
	resp, _ := client.Post()
	fmt.Println(string(resp.GetBody()))
	resp, _ = client.Post()
	fmt.Println(string(resp.GetBody()))
	client.SetUrl("http://10.96.114.84/add.php")
	client.SetHost("phpmyadmin.wendao.com")
	resp, _ = client.Post()
	fmt.Println(string(resp.GetBody()))
}

//直接设置请求的POST的body信息，没有字段，没有文件
func TestHttpClient_SetRawPostBody(t *testing.T) {
	url := "http://10.96.114.84/add.php"
	client := NewHttpClient(url, context.Background())
	client.AddCookie("c1", "v1")
	client.SetUserAgent(testUserAgent)
	client.SetReferer("http://baidu.com")
	client.AddHeader("myHeaderKey", "myHeaderValue")
	client.AddHeader("Content-Type", "application/json")
	client.SetHost("test.wendao.com")
	//设置原始的请求信息
	client.GetBuffer().Write([]byte("laskdfjalksd;fjals;djal;dsfjalds;kjadlsf;jk"))
	resp, err := client.Post()
	fmt.Println(err)
	fmt.Println(string(resp.GetBody()))
	resp, err = client.Post()
	fmt.Println(err)
	fmt.Println(string(resp.GetBody()))
	resp, err = client.Post()
	fmt.Println(err)
	fmt.Println(string(resp.GetBody()))
}
