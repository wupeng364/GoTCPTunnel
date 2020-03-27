// Copyright (C) 2019 WuPeng <wupeng364@outlook.com>.
// Use of this source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package tokentool

import (
	"strconv"
	"testing"
	"time"
)

// 测试令牌生成器
func TestTokenManager(t *testing.T) {
	count := 100000
	timeStart := time.Now().Nanosecond()
	tokenManager := (&TokenManager{}).Init()
	tokens := make([]string, count)
	// 测试申请令牌&获取令牌
	for i := 0; i < count; i++ {
		if i > 5 && i < 100 {
			tokens[i] = tokenManager.AskToken(i, int64(i))
		} else {
			tokens[i] = tokenManager.AskToken(i, -1)
		}
		// 立刻读取
		_, ok := tokenManager.GetTokenBody(tokens[i])
		if !ok {
			t.Log("无法查询申请的令牌")
			t.FailNow()
		}
	}
	tokensNew := tokenManager.ListTokens()
	timeE := (time.Now().Nanosecond() - timeStart) / 1000000
	t.Logf("计划个数: %d, 实际记录个数: %d, 花费时间: %d Millisecond", count, len(tokensNew), timeE)
	if len(tokensNew) != count {
		t.FailNow()
	}
	// 测试令牌过期
	for i := 0; i < count; i++ {
		if i > 5 && i < 100 {
			tokenManager.RefreshToken(tokens[i])
		}
		// 立刻读取
		_, ok := tokenManager.GetTokenBody(tokens[i])
		if !ok {
			t.Log("令牌提前过期")
			t.FailNow()
		}
	}
	time.Sleep(time.Duration(10) * time.Second)
	tokensNew = tokenManager.ListTokens()
	timeE = (time.Now().Nanosecond() - timeStart) / 1000000
	t.Logf("%d Millisecond后, 计划记录个数: %d, 实际记录个数: %d", timeE, count, len(tokensNew))
	if len(tokensNew) == count {
		t.Log("令牌提未过期")
		t.FailNow()
	}
	// 令牌清空
	tokenManager.Clear()
	tokensNew = tokenManager.ListTokens()
	if len(tokensNew) > 0 {
		t.Log("令牌无法清空")
		t.FailNow()
	}
	//
	type testStruct struct {
		body1 string
	}
	token := tokenManager.AskToken(&testStruct{body1: "1"}, -1)
	tokenbody, _ := tokenManager.GetTokenBody(token)
	test := tokenbody.(*testStruct)
	t.Log(test)
	test.body1 = "2"
	tokenbody, _ = tokenManager.GetTokenBody(token)
	test = tokenbody.(*testStruct)
	t.Log(test.body1)
	//
	tokenManager.Destroy()
}

// 缓存模块
func TestCacheManager(t *testing.T) {
	clib := "_test_"
	exp := int64(5)
	count := 100000
	cachemanager := (&CacheManager{}).Init()
	err := cachemanager.RegLib(clib, exp)
	if nil != err {
		t.Log(err)
		t.FailNow()
	}
	// 赋值
	for i := 0; i < count; i++ {
		key := "key_" + strconv.FormatInt(int64(i), 10)
		err = cachemanager.Set(clib, key, i)
		if nil != err {
			t.Log(err)
			t.FailNow()
		}
		_, ok := cachemanager.Get(clib, key)
		if !ok {
			t.Log("无法查询到缓存")
			t.FailNow()
		}
	}
	keys := cachemanager.Keys(clib)
	if len(keys) == 0 {
		t.Log("无法列出所有key")
		t.FailNow()
	}
	// 取值
	time.Sleep(time.Duration(5) * time.Second)
	for i := 0; i < count; i++ {
		key := "key_" + strconv.FormatInt(int64(i), 10)
		_, ok := cachemanager.Get(clib, key)
		if ok {
			t.Log("缓存未销毁")
			t.FailNow()
		}
	}
}

// 令牌生成器
func BenchmarkTokenManager(t *testing.B) {
	count := 5
	timeStart := time.Now().Nanosecond()
	tokenManager := (&TokenManager{}).Init()
	tokens := make([]string, count)
	// 测试申请令牌&获取令牌
	for i := 0; i < count; i++ {
		tokens[i] = tokenManager.AskToken(i, int64(i))
	}
	// list
	tokensNew := tokenManager.ListTokens()
	timeE := (time.Now().Nanosecond() - timeStart) / 1000000
	t.Logf("计划个数: %d, 实际记录个数: %d, 花费时间: %d Millisecond", count, len(tokensNew), timeE)
	// 测试令牌过期
	for i := 0; i < count; i++ {
		tokenManager.RefreshToken(tokens[i])
	}
	tokensNew = tokenManager.ListTokens()
	timeE = (time.Now().Nanosecond() - timeStart) / 1000000
	t.Logf("%d Millisecond后, 计划记录个数: %d, 实际记录个数: %d", timeE, count, len(tokensNew))
	// 令牌清空
	tokenManager.Clear()
	tokensNew = tokenManager.ListTokens()
	if len(tokensNew) > 0 {
		t.Log("令牌无法清空")
		t.FailNow()
	}
	//
	tokenManager.Destroy()
}

// 缓存模块
func BenchmarkCacheManager(t *testing.B) {
	clib := "_test_"
	exp := int64(5)
	count := 5
	var err error
	cachemanager := (&CacheManager{}).Init()
	for i := 0; i < count; i++ {
		si := strconv.FormatInt(int64(i), 10)
		err = cachemanager.RegLib(clib+si, exp+int64(i))
		if nil != err {
			t.Log(err)
			t.FailNow()
		}
	}
	// 赋值
	for i := 0; i < count; i++ {
		si := strconv.FormatInt(int64(i), 10)
		err = cachemanager.Set(clib+si, si, i)
		if nil != err {
			t.Log(err)
			t.FailNow()
		}
		_, ok := cachemanager.Get(clib+si, si)
		if !ok {
			t.Log("无法查询到缓存")
			t.FailNow()
		}
		keys := cachemanager.Keys(clib + si)
		if len(keys) == 0 {
			t.Log("无法列出所有key")
			t.FailNow()
		}
	}
}
