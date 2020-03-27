// Copyright (C) 2019 WuPeng <wupeng364@outlook.com>.
// Use of mloader source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of mloader software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and mloader permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package mloader

import (
	"errors"
	"gutils/conftool"
	"gutils/fstool"
	"gutils/strtool"
)

// NewAsJSONRecorder 使用JSON文件的方式记录模块信息
// 此方法依赖了 conftool, strtool, fstool 包
func NewAsJSONRecorder(savePath string) (*Loader, error) {
	// 实例化模块加载器
	recorder := &jsonrecorder{}
	err := recorder.init(savePath)
	if nil != err {
		return nil, err
	}
	return New(recorder), nil
}

// 模块记录器可可以自己实现, 可以存储在网络配置服务器上
type jsonrecorder struct {
	config *conftool.JSONCFG
}

// 实例化
func (jrd *jsonrecorder) init(savePath string) error {
	if len(savePath) == 0 {
		return errors.New("path is empty")
	}
	// 创建父级目录
	parent := strtool.GetPathParent(savePath)
	if !fstool.IsExist(parent) {
		err := fstool.MkdirAll(parent)
		if nil != err {
			return err
		}
	}
	jrd.config = &conftool.JSONCFG{}
	return jrd.config.InitConfig(savePath)
}

// 读取配置
func (jrd *jsonrecorder) GetValue(key string) string {
	return jrd.config.GetConfig(key).ToString("")
}

// 写入配置
func (jrd *jsonrecorder) SetValue(key string, value string) error {
	return jrd.config.SetConfig(key, value)
}
