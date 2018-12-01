// 重新封装的http请求类
// 此处是从go-http里扒过来的，由于go-http依赖的库太多了，broker/fc-collector都不敢贸然全部更新
// 所以，此库是go-http的简化版本，不支持go-pool、disf等特性，只是简单的发起http请求
// 此处操作post数据时主要用了buf缓冲空间，并提供了直接操作此buffer的方法
// 理论上，日常使用足够了，特殊需求请自行添加相应的方法及设置
//
// @author      Liu Yongshuai<liuyongshuai@hotmail.com>
// @date        2018-10-29 19:38

package goUtils

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//构造一个请求结构体
func NewHttpClient(httpUrl string, ctx context.Context) *HttpClient {
	ret := &HttpClient{
		hClient: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: false,                                 //默认长链接
				TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, //忽略证书校验
			},
		},
		url:         httpUrl,
		ctx:         ctx,                                          //上下文信息，自己传进来
		timeout:     time.Duration(int64(3) * int64(time.Second)), //默认3秒
		retry:       3,                                            //默认重试三次
		headers:     make(http.Header),                            //头信息，所有的请求都要
		buf:         new(bytes.Buffer),                            //最终body缓冲区，一般用于post/put
		vals:        make(url.Values),                             //用于PostForm的KV列表
		uploadFiles: []HttpUploadFile{},                           //上传文件的列表
	}
	return ret
}

//请求结构体
type HttpClient struct {
	hClient     *http.Client     //http.Client
	url         string           //请求的URL
	buf         *bytes.Buffer    //发送的数据，POST用
	vals        url.Values       //提交上来的数据，POST用
	headers     http.Header      //header信息
	uploadFiles []HttpUploadFile //上传的文件列表
	timeout     time.Duration    //整体超时时间。
	retry       int              //重试次数。
	keepalive   bool             //是否保持连接
	proxy       string           //代理地理
	ctx         context.Context  //需要的上下文信息
}

//上传文件的设置
type HttpUploadFile struct {
	FieldName string //上传文件时用的字段名称
	FilePath  string //文件的绝对路径
	FileName  string //上传时显示的文件名称，如果为空则取filePath的basename
}

//设置整体超时时间，默认3秒
func (ehc *HttpClient) SetTimeout(t time.Duration) *HttpClient {
	ehc.timeout = t
	ehc.hClient.Timeout = t
	return ehc
}

//设置重试次数，默认3次
func (ehc *HttpClient) SetRetryTimes(t int) *HttpClient {
	ehc.retry = t
	return ehc
}

//添加要上传的文件
func (ehc *HttpClient) AddFile(fieldName, filePath, fileName string) *HttpClient {
	if !FileExists(filePath) {
		LogErrorf("_soda_eslib_http_AddFile_out||file=%v||file not exists", filePath)
		return ehc
	}
	if len(fileName) <= 0 {
		fileName = filepath.Base(filePath)
	}
	ehc.uploadFiles = append(ehc.uploadFiles, HttpUploadFile{
		FieldName: fieldName,
		FilePath:  filePath,
		FileName:  fileName,
	})
	return ehc
}

//添加单条header信息
func (ehc *HttpClient) AddHeader(k string, v string) *HttpClient {
	ehc.headers.Add(k, v)
	return ehc
}

//添加单条header信息
func (ehc *HttpClient) SetHeader(k string, v string) *HttpClient {
	ehc.headers.Set(k, v)
	return ehc
}

//批量设置头信息
func (ehc *HttpClient) AddHeaders(hs map[string]string) *HttpClient {
	if hs == nil {
		return ehc
	}
	for k, v := range hs {
		ehc.AddHeader(k, v)
	}
	return ehc
}

//批量设置头信息
func (ehc *HttpClient) SetHeaders(hs map[string]string) *HttpClient {
	if hs == nil {
		return ehc
	}
	for k, v := range hs {
		ehc.SetHeader(k, v)
	}
	return ehc
}

//设置要请求的host（设置header的相应值）
func (ehc *HttpClient) SetHost(host string) *HttpClient {
	ehc.SetHeader("Host", host)
	return ehc
}

//设置URL
func (ehc *HttpClient) SetUrl(u string) *HttpClient {
	ehc.url = u
	return ehc
}

//设置长连接选项
func (ehc *HttpClient) SetKeepAlive(b bool) *HttpClient {
	ehc.hClient.Transport.(*http.Transport).DisableKeepAlives = !b
	return ehc
}

//设置代理用的地址和端口
func (ehc *HttpClient) SetProxy(proxyHost string) *HttpClient {
	//如果只给了IP:PORT这样的，默认为http方式
	check, _ := regexp.MatchString(`^[\d]{1,3}\.[\d]{1,3}\.[\d]{1,3}\.[\d]{1,3}:[\d]{1,5}$`, proxyHost)
	if check {
		proxyHost = "http://" + proxyHost
	}
	ehc.hClient.Transport.(*http.Transport).Proxy = func(_ *http.Request) (*url.URL, error) {
		return url.Parse(proxyHost)
	}
	return ehc
}

//设置userAgent（设置header的相应值）
func (ehc *HttpClient) SetUserAgent(ua string) *HttpClient {
	ehc.SetHeader("User-Agent", ua)
	return ehc
}

//设置cookie（设置header的相应值）
func (ehc *HttpClient) SetRawCookie(ck string) *HttpClient {
	ehc.SetHeader("Cookie", ck)
	return ehc
}

//设置内容类型（设置header的相应值）
func (ehc *HttpClient) SetContentType(ct string) *HttpClient {
	ehc.SetHeader("Content-Type", ct)
	return ehc
}

//设置json内容类型（设置header的相应值）
func (ehc *HttpClient) SetContentTypeJson() *HttpClient {
	ehc.SetHeader("Content-Type", "application/json")
	return ehc
}

//设置二进制流内容类型（设置header的相应值）
func (ehc *HttpClient) SetContentTypeOctetStream() *HttpClient {
	ehc.SetHeader("Content-Type", "application/octet-stream")
	return ehc
}

//设置表单内容类型（设置header的相应值）
func (ehc *HttpClient) SetContentTypeFormUrlEncoded() *HttpClient {
	ehc.SetHeader("Content-Type", "application/x-www-form-urlencoded")
	return ehc
}

//设置为Ajax请求（设置header的相应值）
func (ehc *HttpClient) SetAjax() *HttpClient {
	ehc.SetHeader("X-Requested-With", "XMLHttpRequest")
	return ehc
}

//从header中读取内容类型
func (ehc *HttpClient) GetContentType() string {
	contentType := ehc.headers.Get("Content-Type")
	if len(contentType) > 0 {
		return contentType
	}
	contentType = ehc.headers.Get("content-type")
	if len(contentType) > 0 {
		return contentType
	}
	return ""
}

//添加单个cookie的键值
func (ehc *HttpClient) AddCookie(k, v string) *HttpClient {
	ehc.AddCookies(map[string]string{k: v})
	return ehc
}

//批量添加cookie的键值
func (ehc *HttpClient) AddCookies(ck map[string]string) *HttpClient {
	if ck == nil {
		return ehc
	}
	if len(ck) <= 0 {
		return ehc
	}
	cks := ehc.headers.Get("Cookie")
	if len(cks) == 0 {
		cks = ehc.headers.Get("cookie")
	}
	kvs := SplitRawCookie(cks)
	for k, v := range ck {
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if len(k) <= 0 {
			continue
		}
		kvs[k] = v
	}
	rck := JoinRawCookie(kvs)
	ehc.SetRawCookie(rck)
	return ehc
}

//设置referer（设置header的相应值）
func (ehc *HttpClient) SetReferer(referer string) *HttpClient {
	ehc.SetHeader("Referer", referer)
	return ehc
}

//设置HTTP Basic Authentication the provided username and password（设置header的相应值）
func (ehc *HttpClient) SetBasicAuth(username, password string) *HttpClient {
	auth := username + ":" + password
	ehc.SetHeader("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
	return ehc
}

//批量添加字段，一般是 PostForm 用，在即上传文件又有字段时Post也用
func (ehc *HttpClient) AddFields(data map[string]string) *HttpClient {
	if len(data) > 0 {
		for k, v := range data {
			ehc.vals.Set(k, v)
		}
	}
	return ehc
}

//添加单个字段，POST用
func (ehc *HttpClient) AddField(k, v string) *HttpClient {
	ehc.vals.Set(k, v)
	return ehc
}

//获取buf地址，直接操作即可。此为POST 操作里的 body
//如传递raw post的body信息时：httpClient.GetBuffer().Write(c)
func (ehc *HttpClient) GetBuffer() *bytes.Buffer {
	return ehc.buf
}

//设置原始的请求的body信息，一般POST/PUT用得着，等同于httpClient.GetBuffer().Write(c)
func (ehc *HttpClient) SetRawRequestBody(b []byte) *HttpClient {
	ehc.buf.Write(b)
	return ehc
}

//发起GET请求并返回数据
func (ehc *HttpClient) Get() (HttpResponse, error) {
	httpReq, err := http.NewRequest("GET", ehc.url, nil)
	if err != nil {
		LogErrorf("_soda_eslib_http_Get_out||%v||err=%v||http.NewRequest failed", ehc.getComErrMsg(), err)
		return HttpResponse{}, err
	}
	response, err := ehc.do(httpReq)
	return ehc.processResponse(response, err)
}

//发起POST请求并返回数据，一般没有上传文件
func (ehc *HttpClient) PostForm() (HttpResponse, error) {
	ehc.SetContentTypeFormUrlEncoded()
	return ehc.Post()
}

//发起head请求并返回数据
func (ehc *HttpClient) Head() (HttpResponse, error) {
	httpReq, err := http.NewRequest("HEAD", ehc.url, nil)
	if err != nil {
		err = fmt.Errorf("_soda_eslib_http_Head_out||%v||err=%v||http.NewRequest failed", ehc.getComErrMsg(), err)
		LogError(err)
		return HttpResponse{}, err
	}
	resp, err := ehc.do(httpReq)
	LogInfof("_soda_eslib_http_Head_out||%v||err=%v", ehc.getComErrMsg(), err)
	return ehc.processResponse(resp, err)
}

//发起POST请求并返回数据，有字段、上传文件，或者raw post用的
func (ehc *HttpClient) Post() (HttpResponse, error) {
	//如果buf为空，要用KV值、上传的文件填充buf，否则就是要POST的raw body
	//因为writer在关闭时会在数据的尾部加上一串东西
	var writer *multipart.Writer
	if ehc.buf.Len() <= 0 {
		writer = ehc.procWriter()
	}
	//sodaClient.Do()这个方法会把请求的body给干没
	//此处传一个临时的buf进去，保留原始的buf信息
	tmpBuf := bytes.Buffer{}
	tmpBuf.Write(ehc.buf.Bytes())
	//注意，此处传进去的tmpBuf经过sodaClient.Do()处理后就没了
	httpReq, err := http.NewRequest("POST", ehc.url, &tmpBuf)
	if err != nil {
		err = fmt.Errorf("_soda_eslib_http_Post_out||%v||err=%v||http.NewRequest failed", ehc.getComErrMsg(), err)
		LogError(err)
		return HttpResponse{}, err
	}
	//提取content-type类型，如果为空要设置值
	contentType := ehc.GetContentType()
	if len(contentType) <= 0 && writer != nil {
		ehc.SetContentType(writer.FormDataContentType())
	}
	//调用go-http的相关方法来处理
	resp, err := ehc.do(httpReq)
	LogInfof("_soda_eslib_http_Post_out||%v||err=%v", ehc.getComErrMsg(), err)
	return ehc.processResponse(resp, err)
}

//往缓冲里写数据，包括上传的文件、提交的字段列表
func (ehc *HttpClient) procWriter() *multipart.Writer {
	writer := multipart.NewWriter(ehc.buf)
	commErrMsg := ehc.getComErrMsg()
	//写入上传的文件
	for _, f := range ehc.uploadFiles {
		formFile, err := writer.CreateFormFile(f.FieldName, f.FileName)
		if err != nil {
			LogErrorf("_soda_eslib_http_AddFile_out||%v||file=%v||err=%v||CreateFormFile failed", commErrMsg, f.FilePath, err)
			continue
		}
		srcFile, err := os.Open(f.FilePath)
		if err != nil {
			LogErrorf("_soda_eslib_http_AddFile_out||%v||err=%v||file=%v||open file failed", commErrMsg, err, f.FilePath)
			continue
		}
		_, err = io.Copy(formFile, srcFile)
		if err != nil {
			LogErrorf("_soda_eslib_http_AddFile_out||%v||err=%v||file=%v||copy file failed", commErrMsg, err, f.FilePath)
		}
		err = srcFile.Close()
		if err != nil {
			LogErrorf("_soda_eslib_http_AddFile_out||%v||err=%v||file=%v||close file failed", commErrMsg, err, f.FilePath)
		}
	}
	//写入POST的值
	for k, vs := range ehc.vals {
		if len(k) <= 0 {
			continue
		}
		err := writer.WriteField(k, vs[0])
		if err != nil {
			LogErrorf("_soda_eslib_http_AddFile_out||%v||err=%v||WriteField failed", commErrMsg, err)
		}
	}
	err := writer.Close()
	if err != nil {
		LogErrorf("_soda_eslib_http_AddFile_out||%v||err=%v||Writer Close failed", commErrMsg, err)
	}
	return writer
}

//获取通用的日志信息
func (ehc *HttpClient) getComErrMsg() string {
	var vals, headers []string
	for k, v := range ehc.vals {
		vals = append(vals, fmt.Sprintf("%v=%v", k, v))
	}
	for k, v := range ehc.headers {
		headers = append(headers, fmt.Sprintf("%v=%v", k, v))
	}
	commErrMsg := fmt.Sprintf("url=%v||localIP=%v||timeout=%v||retry=%v||vals: %v||headers: %v",
		ehc.url, LocalHostIP, ehc.timeout, ehc.retry, strings.Join(vals, ", "), strings.Join(headers, ", "))
	return commErrMsg
}

//处理响应信息
func (ehc *HttpClient) processResponse(response *http.Response, err error) (HttpResponse, error) {
	ret := HttpResponse{header: make(map[string]string)}
	if response == nil {
		return ret, err
	}
	commErrMsg := ehc.getComErrMsg()
	if err != nil {
		LogErrorf("_soda_eslib_http_processResponse_out||%v||err=%v", commErrMsg, err)
		return ret, err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			LogErrorf("_soda_eslib_http_processResponse_out||%v||err=%v||response.Body.Close() failed", commErrMsg, err)
		}
	}()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		ret.err = err
		LogErrorf("_soda_eslib_http_processResponse_out||%v||err=%v||read resp body failed", commErrMsg, err)
		return ret, err
	}
	ret.body = body
	for k, vs := range response.Header {
		ret.header[k] = strings.Join(vs, ";")
	}
	for _, cptr := range response.Cookies() {
		ret.setCookie = append(ret.setCookie, *cptr)
	}
	ret.transferEncoding = response.TransferEncoding
	ret.status = response.Status
	ret.statusCode = response.StatusCode
	ret.proto = response.Proto
	ret.contentLen = response.ContentLength
	if strings.HasPrefix(ret.status, "3") {
		loc, err := response.Location()
		if err == nil && loc != nil {
			ret.location = loc.String()
		} else {
			ret.err = err
			LogErrorf("_soda_eslib_http_processResponse_out||%v||err=%v||read resp location", commErrMsg, err)
		}
	}
	return ret, nil
}

//执行请求操作，从go-http里扒过来的方法
func (ehc *HttpClient) do(httpReq *http.Request) (resp *http.Response, err error) {
	httpReq.URL, _ = url.Parse(ehc.url)
	//设置请求的header信息
	for k, v := range ehc.headers {
		if len(k) <= 0 || len(v) <= 0 {
			continue
		}
		httpReq.Header.Set(k, v[0])
	}

	var body []byte
	var nextBody io.ReadCloser
	var retry int
	var statusCode int
	startTime := time.Now().UnixNano()
	defer func() {
		procTime := ProcTime(startTime, time.Now().UnixNano())
		var status, errmsg, code string
		if err == nil {
			status = "success"
			if statusCode != 0 && statusCode != http.StatusOK {
				code = strconv.Itoa(statusCode)
			}
		} else {
			status = "failure"
			code = "error"
			errmsg = fmt.Sprintf("||errmsg=%v", err)
		}
		LogInfof("status=%v||proc_time=%f||code=%v%v", status, procTime, code, errmsg)
	}()

	//重试次数
	totalRetryTimes := ehc.retry
	if totalRetryTimes < 0 {
		totalRetryTimes = 0
	}

	//开始请求
	for ; retry <= totalRetryTimes; retry++ {
		resp = nil
		err = ehc.ctx.Err()
		if err != nil {
			LogErrorf("_soda_eslib_httpClient_error||%v||err=ctx.Err||msg=%v", ehc.getComErrMsg(), err)
			return
		}

		// 重新计算 deadline。
		if deadline, ok := ehc.ctx.Deadline(); ok {
			delta := deadline.Sub(time.Now())
			if delta <= 0 {
				err = context.DeadlineExceeded
				return
			}
			if delta > ehc.timeout {
				delta = ehc.timeout
			}
			ehc.hClient.Timeout = delta
		} else {
			ehc.hClient.Timeout = ehc.timeout
		}

		// 如果可能需要重试，必须保留 Body 的内容用于下次重试时候使用。
		if httpReq.Body != nil {
			if totalRetryTimes > 0 {
				if httpReq.GetBody != nil {
					var e error
					nextBody, e = httpReq.GetBody()
					if e != nil {
						nextBody = nil
					}
				}

				if nextBody == nil {
					if body == nil {
						buf := &bytes.Buffer{}
						io.Copy(buf, httpReq.Body)
						body = buf.Bytes()
					}
					nextBody = ioutil.NopCloser(bytes.NewBuffer(body))
				}
			}
		} else {
			if nextBody != nil {
				httpReq.Body = nextBody
			}
		}
		httpReq.Header.Set("custom-header-traceid", FakeTraceId())

		//开始请求http接口
		resp, err = ehc.hClient.Do(httpReq)

		// 重置 Body，这样下次就能够使用缓存 nextBody 来发送请求。
		httpReq.Body = nil

		// 根据官方文档，仅当 err != nil 的时候是可以自动重试，其他情况下都不应该重试。
		if err != nil {
			LogErrorf("_soda_eslib_httpClient_error||%v||err=fail to send request||msg=%v", ehc.getComErrMsg(), err)
			continue
		}

		break
	}

	if resp != nil {
		statusCode = resp.StatusCode
	}

	if resp == nil && err == nil {
		if totalRetryTimes > 0 {
			err = errors.New("too many times to retry")
		} else {
			err = errors.New("fail to send request")
		}
	}

	return
}