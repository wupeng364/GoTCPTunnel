// Copyright (C) 2019 WuPeng <wupeng364@outlook.com>.
// Use of this source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// HTTP客户端工具

package hctool

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultTimeout = 30 // 单位s
)

// BuildURLWithMap 使用Map结构构件url请求参数
// {key:value} => /url/xxx?key=value
func BuildURLWithMap(url string, params map[string]string) string {
	result := url
	lenP := len(params)
	if params != nil && lenP > 0 {
		result += "?"
		for key, val := range params {
			result += key + "=" + val
			if lenP > 1 {
				result += "&"
			}
			lenP--
		}
	}
	return result
}

// BuildURLWithArray 使用二维数组结构构件url请求参数
// [[key,value], ...] => /url/xxx?key=value
func BuildURLWithArray(url string, params [][]string) string {
	result := url
	lenP := len(params)
	if params != nil && lenP > 0 {
		result += "?"
		for i := 0; i < lenP; i++ {
			if len(params[i]) >= 2 {
				result += params[i][0] + "=" + params[i][1]
				if i < lenP-1 {
					result += "&"
				}
			}
		}
	}
	return result
}

// Get Get请求
func Get(url string, params map[string]string, headers map[string]string) (*http.Response, error) {
	return DoRequest("GET", url, params, headers, defaultTimeout)
}

// GetBodyStr Get请求, 返回请求结果内容字符串
func GetBodyStr(url string, params map[string]string, headers map[string]string) (string, error) {
	resp, err := Get(url, params, headers)
	defer func() {
		if nil != resp {
			resp.Body.Close()
		}
	}()
	if nil != err {
		return "", err
	}
	// red response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// Post Post请求
func Post(url string, params map[string]string, headers map[string]string) (*http.Response, error) {
	return DoRequest("POST", url, params, headers, defaultTimeout)
}

// DoRequest 发送请求(请求方式, url, 参数, 头信息, 超时设置)
// 默认使用 application/x-www-form-urlencoded 方式发送请求, 可通过头信息覆盖
func DoRequest(reqType, url string, params, headers map[string]string, timeout int64) (*http.Response, error) {

	// build query
	query := ""
	lenP := len(params)
	if params != nil && lenP > 0 {
		for key, val := range params {
			query += key + "=" + val
			if lenP > 1 {
				query += "&"
			}
			lenP--
		}
	}

	// build request method
	req, err := http.NewRequest(reqType, url, strings.NewReader(query))
	if err != nil {
		return nil, err
	}

	// set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if headers != nil && len(headers) > 0 {
		for key, val := range headers {
			req.Header.Set(key, val)
		}
	}

	// do request
	client := &http.Client{}
	if timeout > -1 {
		client.Timeout = time.Second * time.Duration(timeout)
	}

	return client.Do(req)
}

// PostJSON 通过Post Json 内容发送请求
func PostJSON(url string, params interface{}, headers map[string]string, timeout int64) (*http.Response, error) {

	// build query
	query, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	// build request method
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(query))
	if err != nil {
		return nil, err
	}

	// set headers
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	if headers != nil && len(headers) > 0 {
		for key, val := range headers {
			req.Header.Set(key, val)
		}
	}

	// do request
	client := &http.Client{}
	if timeout > -1 {
		client.Timeout = time.Second * time.Duration(timeout)
	}
	return client.Do(req)
}

// PostFile 发送文件使用默认的file表单字段&无附加头信息 (url, 本地路径)
func PostFile(url, filePath string) (*http.Response, error) {
	return PostMultiFile(url, filePath, nil, "")
}

// PostMultiFile 通过Form表单提交文件
func PostMultiFile(url, filePath string, headers map[string]string, paramName string) (*http.Response, error) {
	bodyBuf := bytes.NewBufferString("")
	bodyWriter := multipart.NewWriter(bodyBuf)
	if len(paramName) == 0 {
		paramName = "file"
	}
	_, err := bodyWriter.CreateFormFile(paramName, filePath)
	if err != nil {
		return nil, err
	}

	fh, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	boundary := bodyWriter.Boundary()
	closeBuf := bytes.NewBufferString("\r\n--" + boundary + "--\r\n")

	reqReader := io.MultiReader(bodyBuf, fh, closeBuf)
	fi, err := fh.Stat()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, reqReader)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "multipart/form-data; boundary="+boundary)
	if headers != nil && len(headers) > 0 {
		for key, val := range headers {
			req.Header.Set(key, val)
		}
	}
	req.ContentLength = fi.Size() + int64(bodyBuf.Len()) + int64(closeBuf.Len())

	client := &http.Client{}
	return client.Do(req)
}
