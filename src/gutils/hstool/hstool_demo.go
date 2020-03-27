// Copyright (C) 2019 WuPeng <wupeng364@outlook.com>.
// Use of this source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package hstool

import (
	"errors"
	"net/http"
)

// StartService 注册路由,并启动服务
// 此函数为ServiceRouter简版的使用方式, 可以根据ServiceRouter自己实现
// 默认读写都不设置超时
func StartService(addr string, handlersMap map[string]HandlersFunc, defaultHandler HandlersFunc) error {
	if handlersMap == nil || len(handlersMap) == 0 {
		return errors.New("handlersMap is nil")
	}
	if addr == "" {
		return errors.New("addr is nil or empty")
	}
	router := &ServiceRouter{}
	router.isDebug = true
	router.defaultHandler = defaultHandler
	router.AddHandlers(handlersMap)

	server := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    0,
		WriteTimeout:   0,
		MaxHeaderBytes: 1 << 20,
	}
	return server.ListenAndServe()
}
