// Copyright (C) 2019 WuPeng <wupeng364@outlook.com>.
// Use of this source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// 缓存工具

package tokentool

import (
	"errors"
	"sync"
)

// CacheManager 基于TokenManager实现的缓存管理器
// 使用前需要调用 init 方法
type CacheManager struct {
	// defaultlib string
	libexp   map[string]int64
	clibs    map[string]*TokenManager
	clibLock *sync.RWMutex
}

// Init 初始化缓存管理器, 一个对象只能初始化一次
func (cm *CacheManager) Init() *CacheManager {
	if nil != cm.clibs {
		return cm
	}
	// cm.defaultlib = "_d_"
	cm.libexp = make(map[string]int64)
	cm.clibs = make(map[string]*TokenManager)
	cm.clibLock = new(sync.RWMutex)
	// 初始默认库, 默认初始化一个名为_d_的缓存库, 该库存储的内容永不过期
	// cm.clibLock.Lock()
	// cm.clibs[cm.defaultlib] = (&TokenManager{}).Init()
	// cm.libexp[cm.defaultlib] = -1
	// cm.clibLock.Unlock()
	return cm
}

// RegLib 注册缓存库
// lib为库名, second:过期时间-1为不过期
func (cm *CacheManager) RegLib(lib string, second int64) error {
	if len(lib) == 0 {
		return errors.New("lib is empty")
	}
	defer cm.clibLock.Unlock()
	cm.clibLock.Lock()

	if _, ok := cm.clibs[lib]; ok {
		return errors.New("lib is exist")
	}
	cm.clibs[lib] = (&TokenManager{}).Init()
	cm.libexp[lib] = second

	return nil
}

// Set 向lib库中设置键为key的值tb
func (cm *CacheManager) Set(lib string, key string, tb interface{}) error {
	defer cm.clibLock.Unlock()
	cm.clibLock.Lock()
	if len(lib) == 0 {
		// lib = cm.defaultlib
		return errors.New("lib is empty")
	}
	tm, ok := cm.clibs[lib]
	if !ok {
		return errors.New("lib not exist")
	}
	lx, ok := cm.libexp[lib]
	if !ok {
		delete(cm.clibs, lib)
		return errors.New("lib not exist")
	}

	tm.PutTokenBody(key, tb, lx)
	return nil
}

// Get 读取缓存信息
func (cm *CacheManager) Get(clib string, key string) (interface{}, bool) {
	defer cm.clibLock.RUnlock()
	cm.clibLock.RLock()

	tm, ok := cm.clibs[clib]
	if !ok {
		return nil, false
	}
	return tm.GetTokenBody(key)
}

// Keys 获取库的所有key
func (cm *CacheManager) Keys(clib string) []string {
	defer cm.clibLock.RUnlock()
	cm.clibLock.RLock()

	tm, ok := cm.clibs[clib]
	if !ok {
		return make([]string, 0)
	}
	return tm.ListTokens()
}

// Clear 清空库内容
func (cm *CacheManager) Clear(clib string) {
	defer cm.clibLock.Unlock()
	cm.clibLock.Lock()

	if tm, ok := cm.clibs[clib]; ok {
		tm.Clear()
	}
}
