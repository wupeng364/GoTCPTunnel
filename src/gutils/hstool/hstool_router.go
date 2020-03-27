// Copyright (C) 2019 WuPeng <wupeng364@outlook.com>.
// Use of this source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// http服务器工具-URL路由管理
// 请求处理逻辑: 过滤器 > 路径匹配 > END
// 过滤器优先级: 全匹配url > 正则url > 全局设定 > 无匹配(next)
// 路径处理优先级: 全匹配url > 正则url > 默认设定 > 无匹配(404)

package hstool

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

// HandlersFunc 定义请求处理器
type HandlersFunc func(http.ResponseWriter, *http.Request)

// FilterNext http请求过滤器, next函数用于触发下一步操作, 不执行就不继续处理请求
type FilterNext func()

// FilterFunc http请求过滤器, next函数用于触发下一步操作, 不执行就不继续处理请求
type FilterFunc func(http.ResponseWriter, *http.Request, FilterNext)

// ServiceRouter 实现了http.Server接口的ServeHTTP方法
type ServiceRouter struct {
	isDebug             bool                    // 调试模式可以打印信息
	defaultHandler      HandlersFunc            // 默认的url处理, 可以用于处理静态资源
	urlHandlersMap      map[string]HandlersFunc // url路径全匹配路由表
	regexpHandlersMap   map[string]HandlersFunc // url路径正则配路由表
	regexpHandlersIndex []string                // url路径正则配路由表-索引(用于保存顺序)
	urlFiltersMap       map[string]FilterFunc   // url路径过滤器
	regexpFiltersMap    map[string]FilterFunc   // url路径正则匹配过滤器
	regexpFiltersIndex  []string                // url路径正则匹配过滤器-索引(用于保存顺序)
	filterWhitelist     map[string]string
	globalFileter       FilterFunc
}

// ServiceRouter 根据注册的路由表调用对应的函数
// 优先匹配全url > 正则url > 默认处理器 > 404
func (serviceRouter *ServiceRouter) doHandle(w http.ResponseWriter, r *http.Request) {
	// 如果是url全匹配, 则直接执行hand函数
	if h, ok := serviceRouter.urlHandlersMap[r.URL.Path]; ok {
		if serviceRouter.isDebug {
			fmt.Println("URL.Handler: ", r.URL.Path)
		}
		h(w, r)
		return
	}

	// 如果是url正则检查, 则需要检查正则, 正则为':'后面的字符
	for _, key := range serviceRouter.regexpHandlersIndex {
		symbolIndex := strings.Index(key, ":")
		if symbolIndex == -1 {
			continue
		}
		baseURL := key[:symbolIndex]
		if !strings.HasPrefix(r.URL.Path, baseURL) {
			continue
		}
		if ok, _ := regexp.MatchString(baseURL+key[symbolIndex+1:], r.URL.Path); ok {
			if serviceRouter.isDebug {
				fmt.Println("URL.Handler.Regexp: ", key)
			}
			serviceRouter.regexpHandlersMap[key](w, r)
			return
		}

	}
	// 没有注册的地址, 使用默认处理器
	if serviceRouter.defaultHandler != nil {
		serviceRouter.defaultHandler(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// doFilter 根据注册的过滤器表调用对应的函数
// 优先匹配全url > 正则url > 全局过滤器 > 直接通过
func (serviceRouter *ServiceRouter) doFilter(w http.ResponseWriter, r *http.Request) {
	// 1.1 检擦是否有指定路径的全路径匹配过滤器设定, 优先处理
	if nil != serviceRouter.urlFiltersMap {
		if h, exist := serviceRouter.urlFiltersMap[r.URL.Path]; exist {
			if serviceRouter.isDebug {
				fmt.Println("URL.Filter: ", r.URL.Path)
			}
			h(w, r, func() {
				serviceRouter.doHandle(w, r)
			})
			return
		}
	}
	// 1.2 检擦是否有指定路径的正则匹配过滤器设定, 优先处理
	if nil != serviceRouter.regexpFiltersIndex && len(serviceRouter.regexpFiltersIndex) > 0 {
		for _, key := range serviceRouter.regexpFiltersIndex {
			symbolIndex := strings.Index(key, ":")
			if symbolIndex == -1 {
				continue
			}
			baseURL := key[:symbolIndex]
			if !strings.HasPrefix(r.URL.Path, baseURL) {
				continue
			}

			if ok, _ := regexp.MatchString(key[:symbolIndex]+key[symbolIndex+1:], r.URL.Path); ok {
				if serviceRouter.isDebug {
					fmt.Println("URL.Filter.Regexp: ", key)
				}
				serviceRouter.regexpFiltersMap[key](w, r, func() {
					serviceRouter.doHandle(w, r)
				})
				return
			}
		}
	}
	// 2. 检擦是否有全局过滤器存在, 如果有则执行它
	if nil != serviceRouter.globalFileter {
		serviceRouter.globalFileter(w, r, func() {
			serviceRouter.doHandle(w, r)
		})
		return
	}
	// 3. 啥也没有设定
	serviceRouter.doHandle(w, r)
}

// ServeHTTP 实现http.Server接口的ServeHTTP方法
func (serviceRouter *ServiceRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if serviceRouter.isDebug {
		fmt.Println("URL.Path: ", r.URL.Path)
	}
	// 处理前进行过滤处理
	serviceRouter.doFilter(w, r)
}

// ClearHandlersMap 清空路由表
func (serviceRouter *ServiceRouter) ClearHandlersMap() {
	serviceRouter.urlHandlersMap = make(map[string]HandlersFunc)
	serviceRouter.regexpHandlersMap = make(map[string]HandlersFunc)
	serviceRouter.regexpHandlersIndex = make([]string, 0, 0)
}

// SetDebug 是否输出url请求信息
func (serviceRouter *ServiceRouter) SetDebug(isDebug bool) {
	serviceRouter.isDebug = isDebug
}

// SetDefaultHandler 设置默认相应函数, 当无匹配时触发
func (serviceRouter *ServiceRouter) SetDefaultHandler(defaultHandler HandlersFunc) {
	serviceRouter.defaultHandler = defaultHandler
}

// SetGlobalFilter 设置全局过滤器, 设置后, 如果不调用next函数则不进行下一步处理
// type FilterFunc func(http.ResponseWriter, *http.Request, func( ))
func (serviceRouter *ServiceRouter) SetGlobalFilter(globalFilter FilterFunc) {
	serviceRouter.globalFileter = globalFilter
}

// AddURLFilter 设置url过滤器, 设置后, 如果不调用next函数则不进行下一步处理
// 过滤器有优先调用权, 正则匹配路径有先后顺序
// type FilterFunc func(http.ResponseWriter, *http.Request, func( ))
func (serviceRouter *ServiceRouter) AddURLFilter(url string, filter FilterFunc) {
	if len(url) == 0 {
		return
	}
	if nil == serviceRouter.urlFiltersMap {
		serviceRouter.urlFiltersMap = make(map[string]FilterFunc)
	}
	if nil == serviceRouter.regexpFiltersMap {
		serviceRouter.regexpFiltersMap = make(map[string]FilterFunc)
	}
	if nil == serviceRouter.regexpFiltersIndex {
		serviceRouter.regexpFiltersIndex = make([]string, 0, 0)
	}
	if strings.Index(url, ":") > -1 {
		serviceRouter.regexpFiltersMap[url] = filter
		serviceRouter.regexpFiltersIndex = append(serviceRouter.regexpFiltersIndex, url)
	} else {
		serviceRouter.urlFiltersMap[url] = filter
	}
}

// removeFilterIndex 删除filter索引
func (serviceRouter *ServiceRouter) removeFilterIndex(url string) {
	if len(url) > 0 {
		for i, key := range serviceRouter.regexpFiltersIndex {
			if key == url {
				serviceRouter.regexpFiltersIndex = append(serviceRouter.regexpFiltersIndex[:i], serviceRouter.regexpFiltersIndex[i+i:]...)
				break
			}
		}
	}
}

// RemoveFilter 删除一个过滤器
func (serviceRouter *ServiceRouter) RemoveFilter(url string) {
	if len(url) == 0 {
		return
	}
	if nil != serviceRouter.regexpHandlersMap {
		if _, ok := serviceRouter.regexpFiltersMap[url]; ok {
			delete(serviceRouter.regexpFiltersMap, url)
			serviceRouter.removeFilterIndex(url)
		}
	}
	if nil != serviceRouter.urlFiltersMap {
		if _, ok := serviceRouter.urlFiltersMap[url]; ok {
			delete(serviceRouter.urlFiltersMap, url)
		}
	}
}

// AddHandlers 批量添加handler
// 全匹配和正则匹配分开存放, 正则表达式以':'符号开始, 如: /upload/:\S+
func (serviceRouter *ServiceRouter) AddHandlers(handlersMap map[string]HandlersFunc) {
	if len(handlersMap) == 0 {
		return
	}
	if nil == serviceRouter.regexpHandlersMap {
		serviceRouter.regexpHandlersMap = make(map[string]HandlersFunc)
		serviceRouter.regexpHandlersIndex = make([]string, 0, 0)
	}
	if nil == serviceRouter.urlHandlersMap {
		serviceRouter.urlHandlersMap = make(map[string]HandlersFunc)
	}
	for key, val := range handlersMap {
		if strings.Index(key, ":") > -1 {
			serviceRouter.regexpHandlersMap[key] = val
			serviceRouter.regexpHandlersIndex = append(serviceRouter.regexpHandlersIndex, key)
		} else {
			serviceRouter.urlHandlersMap[key] = val
		}
	}
}

// AddHandler 添加handler
// 全匹配和正则匹配分开存放, 正则表达式以':'符号开始, 如: /upload/:\S+
func (serviceRouter *ServiceRouter) AddHandler(url string, handler HandlersFunc) {
	if len(url) == 0 {
		return
	}
	if nil == serviceRouter.regexpHandlersMap {
		serviceRouter.regexpHandlersMap = make(map[string]HandlersFunc)
		serviceRouter.regexpHandlersIndex = make([]string, 0, 0)
	}
	if nil == serviceRouter.urlHandlersMap {
		serviceRouter.urlHandlersMap = make(map[string]HandlersFunc)
	}
	if strings.Index(url, ":") > -1 {
		serviceRouter.regexpHandlersMap[url] = handler
		serviceRouter.regexpHandlersIndex = append(serviceRouter.regexpHandlersIndex, url)
	} else {
		serviceRouter.urlHandlersMap[url] = handler
	}
}

// removeHandlerIndex 删除handler索引
func (serviceRouter *ServiceRouter) removeHandlerIndex(url string) {
	if len(url) > 0 {
		for i, key := range serviceRouter.regexpHandlersIndex {
			if key == url {
				serviceRouter.regexpHandlersIndex = append(serviceRouter.regexpHandlersIndex[:i], serviceRouter.regexpHandlersIndex[i+i:]...)
				break
			}
		}
	}
}

// RemoveHandler 删除一个路由表
func (serviceRouter *ServiceRouter) RemoveHandler(url string) {
	if len(url) == 0 {
		return
	}
	if nil != serviceRouter.regexpHandlersMap {
		if _, ok := serviceRouter.regexpHandlersMap[url]; ok {
			delete(serviceRouter.regexpHandlersMap, url)
			serviceRouter.removeHandlerIndex(url)
		}
	}
	if nil != serviceRouter.urlHandlersMap {
		if _, ok := serviceRouter.urlHandlersMap[url]; ok {
			delete(serviceRouter.urlHandlersMap, url)
		}
	}
}
