// Copyright (C) 2019 WuPeng <wupeng364@outlook.com>.
// Use of this source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// key-value临时存储工具
// 依赖包: strtool

package tokentool

import (
	"gutils/strtool"
	"sync"
	"time"
)

// TokenManager 令牌管理器, 可实现临时对象的存储
// 使用前需要调用 init 方法
type TokenManager struct {
	destroyed bool // 是否销毁对象
	tokenMap  map[string]tokenObject
	tokenLock *sync.RWMutex
}

// tokenObject 用于保存token内容的对象
type tokenObject struct {
	regtime int64       // 内部计算用-注册时间
	expired int64       // 内部计算用-过期时间
	O       interface{} // token内容
}

// Init 初始化-启动一个管理线程, 负责令牌的生命周期
func (tm *TokenManager) Init() *TokenManager {
	if nil != tm.tokenMap {
		return tm
	}
	tm.destroyed = false
	tm.tokenMap = make(map[string]tokenObject)
	tm.tokenLock = new(sync.RWMutex)

	// 定期清理
	go func() {
		for {
			if tm.destroyed {
				break
			}
			if len(tm.tokenMap) >= 0 {
				tm.tokenLock.Lock() // 读取时锁定map, 防止中途修改
				for key, val := range tm.tokenMap {
					if val.expired == -1 {
						continue
					}
					now := time.Now().UnixNano()
					if val.expired <= now {
						// fmt.Println("remove: ", key, val.expired - now, val)
						delete(tm.tokenMap, key)
					}
				}
				tm.tokenLock.Unlock()
			}
			time.Sleep(time.Duration(1) * time.Second)
		}
		if tm.destroyed {
			tm.tokenMap = nil
			tm.tokenLock.Unlock()
			tm.tokenLock = nil
		}
	}()
	return tm
}

// AskToken 生成令牌, tb:存储内容, second:过期时间, 单位秒
// 当 second=-1时, 不会自动销毁内存中的信息
func (tm *TokenManager) AskToken(tb interface{}, second int64) string {
	token := strtool.GetUUID()
	tm.PutTokenBody(token, tb, second)
	return token
}

// PutTokenBody 设置令牌内容, key: tooken字符, 存在则覆盖, tb:存储内容, second:过期时间, 单位秒
// 当 second=-1时, 不会自动销毁内存中的信息
func (tm *TokenManager) PutTokenBody(token string, tb interface{}, second int64) {
	tkb := tokenObject{
		O:       tb,
		regtime: time.Now().UnixNano(),
	}
	if second > -1 {
		tkb.expired = tkb.regtime + second*int64(time.Second)
	} else {
		tkb.expired = -1
	}
	tm.tokenLock.Lock()
	defer tm.tokenLock.Unlock()
	tm.tokenMap[token] = tkb
}

// GetTokenBody 获取令牌信息
func (tm *TokenManager) GetTokenBody(tk string) (interface{}, bool) {
	tm.tokenLock.RLock()
	defer tm.tokenLock.RUnlock()

	val, ok := tm.tokenMap[tk]
	if ok {
		if val.expired != int64(-1) && val.expired <= time.Now().UnixNano() {
			return nil, false
		}
		return val.O, ok
	}
	return nil, ok
}

// RefreshToken 刷新|重置令牌过期时间
func (tm *TokenManager) RefreshToken(tk string) {
	tm.tokenLock.Lock()
	defer tm.tokenLock.Unlock()

	val, ok := tm.tokenMap[tk]
	if ok {
		now := time.Now().UnixNano()
		if val.expired <= now {
			return
		}
		used := val.expired - val.regtime
		val.regtime = time.Now().UnixNano()
		val.expired = val.regtime + used
		tm.tokenMap[tk] = val
	}
}

// ListTokens 列出所有的token
func (tm *TokenManager) ListTokens() []string {
	defer tm.tokenLock.RUnlock()
	tm.tokenLock.RLock()
	keys := make([]string, 0)
	for key, val := range tm.tokenMap {
		if val.expired == -1 {
			keys = append(keys, key)
			continue
		}
		if val.expired == -1 || val.expired > time.Now().UnixNano() {
			keys = append(keys, key)
		}
	}
	return keys
}

// GetExpiredNano 获取档期那token还有多久过期, 单位纳秒
func (tm *TokenManager) GetExpiredNano(tk string) int64 {
	val, ok := tm.tokenMap[tk]
	if !ok {
		return -1
	}
	return val.expired - val.regtime
}

// Clear 销毁整个对象, 销毁后不能在使用此对象, 需要重新初始化
func (tm *TokenManager) Clear() {
	tm.tokenLock.Lock()
	defer tm.tokenLock.Unlock()
	tm.tokenMap = make(map[string]tokenObject)
}

// DestroyToken 销毁令牌
func (tm *TokenManager) DestroyToken(tk string) {
	tm.tokenLock.Lock()
	defer tm.tokenLock.Unlock()

	if _, ok := tm.tokenMap[tk]; ok {
		delete(tm.tokenMap, tk)
	}
}

// Destroy 销毁整个对象, 销毁后不能在使用此对象, 需要重新初始化
func (tm *TokenManager) Destroy() {
	tm.tokenLock.Lock()
	tm.destroyed = true
}
