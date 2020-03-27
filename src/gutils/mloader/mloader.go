// Copyright (C) 2019 WuPeng <wupeng364@outlook.com>.
// Use of mloader source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of mloader software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and mloader permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// 模块加载器
// 依赖包: types.Object strtool

package mloader

import (
	"errors"
	"fmt"
	"gutils/strtool"
	"gutils/types"
	"reflect"
	"strconv"
	"sync"
	"time"
)

const (
	moduleVersionPrec = 2 // 版本信息保留小数位数
)

// Loader 模块加载器, 实例化后可实现统一管理模板
type Loader struct {
	instanceID string                 // 加载器实例ID
	modules    map[string]interface{} // 模块Map表
	mdslock    *sync.RWMutex          // 对modules对象的读写锁
	mparams    map[string]interface{} // 保存在模块对象中共享的字段key-value
	mpslock    *sync.RWMutex          // 对mparams对象的读写锁
	mrecord    mrecorder              // 模块信息记录器
}

// Opts 模块配置项
type Opts struct {
	Name        string             // 模块ID
	Version     float64            // 模块版本
	Description string             // 模块描述
	OnReady     func(mctx *Loader) // 每次加载模块开始之前执行
	OnSetup     func()             // 模块安装, 一个模块只初始化一次
	OnUpdate    func(cv float64)   // 模块升级, 一个版本执行一次
	OnInit      func()             // 每次模块安装、升级后执行一次
}

// Template 模块模板, 实现这个接口便可加载
type Template interface {
	ModuleOpts() Opts
}

// 用于记录模块信息
type mrecorder interface {
	GetValue(key string) string
	SetValue(key string, value string) error
}

// Returns 函数执行后的返回值, 暂时不封装
type Returns []reflect.Value

// New 实例一个加载器对象
func New(mrecord mrecorder) *Loader {
	res := &Loader{
		mrecord:    mrecord,
		modules:    make(map[string]interface{}),
		mdslock:    new(sync.RWMutex),
		mparams:    make(map[string]interface{}),
		mpslock:    new(sync.RWMutex),
		instanceID: strtool.GetUUID(),
	}
	return res
}

// Loads 初始化模块 - DO Setup -> Check Ver -> Do Init
func (mloader *Loader) Loads(mts ...Template) {
	for _, mt := range mts {
		mloader.Load(mt)
	}
}

// Load 初始化模块 - DO Setup -> Check Ver -> Do Init
func (mloader *Loader) Load(mt Template) {
	opts := mt.ModuleOpts()
	fmt.Printf(">Loading %s(%s)[%p] start \r\n", opts.Name, opts.Description, mt)
	// DO Ready
	mloader.doReady(opts)
	// DO Setup
	mloader.doSetup(opts)
	// Check Ver
	mloader.doCheckVersion(opts)
	// Do Init
	mloader.doInit(opts)
	// Load End
	mloader.doEnd(opts, mt)
	fmt.Printf(">Loading %s complete \r\n", opts.Name)

}

// Invoke 模块调用, 返回 []reflect.Value, 返回值暂时无法处理
func (mloader *Loader) Invoke(name string, method string, params ...interface{}) (Returns, error) {
	mloader.mdslock.RLock()
	defer mloader.mdslock.RUnlock()
	if module, ok := mloader.modules[name]; ok {
		val := reflect.ValueOf(module)
		fun := val.MethodByName(method)
		fmt.Printf("> Invoke: "+name+"."+method+", %v, %+v \r\n", fun, &fun)
		args := make([]reflect.Value, len(params))
		for i, temp := range params {
			args[i] = reflect.ValueOf(temp)
		}
		return fun.Call(args), nil
	}
	return nil, errors.New("module not find: " + name)
}

// GetInstanceID 获取实例的ID
func (mloader *Loader) GetInstanceID() string {
	return mloader.instanceID
}

// SetParam 设置变量, 保存在模板加载器实例内部
func (mloader *Loader) SetParam(key string, val interface{}) {
	mloader.mpslock.Lock()
	defer mloader.mpslock.Unlock()
	mloader.mparams[key] = val
}

// GetParam 模板加载器实例上的变量
func (mloader *Loader) GetParam(key string) types.Object {
	mloader.mpslock.RLock()
	defer mloader.mpslock.RUnlock()
	val, ok := mloader.mparams[key]
	if !ok {
		return types.Object{}
	}
	return types.Object{O: val}
}

// GetModuleByName 根据模块Name获取模块指针记录, 可以获取一个已经实例化的模块
func (mloader *Loader) GetModuleByName(name string) (val interface{}, ok bool) {
	mloader.mdslock.RLock()
	defer mloader.mdslock.RUnlock()
	v, ok := mloader.modules[name]
	return v, ok
}

// GetModuleByTemplate 根据模板对象获取模块指针记录, 可以获取一个已经实例化的模块
func (mloader *Loader) GetModuleByTemplate(mt Template) interface{} {
	mopts := mt.ModuleOpts()
	if val, ok := mloader.GetModuleByName(mopts.Name); ok {
		return val
	}
	panic(errors.New("module not find: " + mopts.Name + "[" + mopts.Description + "]"))
}

// getInstalledVersion 获取模块版本号
func (mloader *Loader) getInstalledVersion(opts Opts) string {
	return mloader.mrecord.GetValue(opts.Name + ".SetupVer")
}

// setVersion 设置模块版本号 - 模块保留小数两位
func (mloader *Loader) setVersion(opts Opts) {
	mloader.mrecord.SetValue(opts.Name+".SetupVer", strconv.FormatFloat(opts.Version, 'f', 2, 64))
	mloader.mrecord.SetValue(opts.Name+".SetupDate", strconv.FormatInt(time.Now().UnixNano(), 10))
}

// doReady 模块准备
func (mloader *Loader) doReady(opts Opts) {
	if nil != opts.OnReady {
		fmt.Printf("  > On ready load \r\n")
		opts.OnReady(mloader)
	}
}

// doSetup 模块安装
func (mloader *Loader) doSetup(opts Opts) {
	if len(mloader.getInstalledVersion(opts)) == 0 {
		if nil != opts.OnSetup {
			fmt.Printf("  > On setup module \r\n")
			opts.OnSetup()
		}

		mloader.setVersion(opts)
	}
}

// doCheckVersion 模块升级
func (mloader *Loader) doCheckVersion(opts Opts) {
	setupVerStr := strconv.FormatFloat(opts.Version, 'f', 2, 64)
	historyVer := mloader.getInstalledVersion(opts)
	if historyVer != setupVerStr {
		if nil != opts.OnUpdate {
			fmt.Printf("  > On update version \r\n")
			hv, err := strconv.ParseFloat(historyVer, 64)
			if nil != err {
				panic(err)
			}
			opts.OnUpdate(hv)
		}

		mloader.setVersion(opts)
	}
}

// doInit 模块初始化
func (mloader *Loader) doInit(opts Opts) {
	if nil != opts.OnInit {
		fmt.Printf("  > On init module \r\n")
		opts.OnInit()
	}
}

// doEnd 模块加载结束
func (mloader *Loader) doEnd(opts Opts, mt Template) {
	mloader.mdslock.Lock()
	defer mloader.mdslock.Unlock()
	mloader.modules[opts.Name] = mt
}
