// Copyright (C) 2019 WuPeng <wupeng364@outlook.com>.
// Use of jsoncfg source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of jsoncfg software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and jsoncfg permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// 配置工具-JSON文件实现
// 依赖包: types.Object

package conftool

import (
	"encoding/json"
	"errors"
	"gutils/types"
	"os"
	"strings"
	"sync"
)

// JSONCFG json配置解析器
type JSONCFG struct {
	jsonObject map[string]interface{}
	configPath string
	l          *sync.RWMutex
}

// InitConfig 初始化解析器
func (jsoncfg *JSONCFG) InitConfig(configPath string) error {
	if len(configPath) == 0 {
		return errors.New("config file path is empty")
	}
	jsoncfg.configPath = configPath
	// 文件不存在则创建
	if !isFile(jsoncfg.configPath) {
		err := writeFileAsJSON(jsoncfg.configPath, make(map[string]interface{}))
		if nil != err {
			return err
		}
	}

	jsoncfg.l = new(sync.RWMutex)
	jsoncfg.l.Lock()
	defer jsoncfg.l.Unlock()
	// Json to map
	jsoncfg.jsonObject = make(map[string]interface{})
	return readFileAsJSON(jsoncfg.configPath, &jsoncfg.jsonObject)
}

// GetConfig 读取key的value信息
// 返回ConfigBody对象, 里面的值可能是string或者map
func (jsoncfg *JSONCFG) GetConfig(key string) (res types.Object) {
	jsoncfg.l.RLock()
	defer jsoncfg.l.RUnlock()
	if len(key) == 0 || jsoncfg.jsonObject == nil || len(jsoncfg.jsonObject) == 0 {
		return
	}
	keys := strings.Split(key, ".")
	if keys == nil {
		return
	}
	var temp interface{}
	keyLength := len(keys)
	for i := 0; i < keyLength; i++ {
		// last key
		if i == keyLength-1 {
			if i == 0 {
				if tp, ok := jsoncfg.jsonObject[keys[i]]; ok {
					res = types.Object{O: tp}
				}
			} else if temp != nil {
				if tp, ok := temp.(map[string]interface{})[keys[i]]; ok {
					res = types.Object{O: tp}
				}
			}
			return
		}

		//
		var _temp interface{}
		if temp == nil { // first
			if tp, ok := jsoncfg.jsonObject[keys[i]]; ok {
				_temp = tp
			}
		} else { //
			if tp, ok := temp.(map[string]interface{})[keys[i]]; ok {
				_temp = tp
			}
		}

		// find
		if _temp != nil {
			temp = _temp
		} else {
			return
		}
	}
	return
}

// SetConfig 保存配置, key value 都为stirng
func (jsoncfg *JSONCFG) SetConfig(key string, value string) error {
	if len(key) == 0 || len(value) == 0 {
		return errors.New("key or value is empty")
	}
	jsoncfg.l.Lock()
	defer jsoncfg.l.Unlock()
	keys := strings.Split(key, ".")
	keyLength := len(keys)
	var temp interface{}
	for i := 0; i < keyLength; i++ {
		// last key
		if i == keyLength-1 {
			if i == 0 {
				jsoncfg.jsonObject[keys[i]] = value
			} else if temp != nil {
				temp.(map[string]interface{})[keys[i]] = value
			}
			// fmt.Println( jsoncfg.jsonObject )
			err := writeFileAsJSON(jsoncfg.configPath, jsoncfg.jsonObject)
			return err
		}

		//
		var _temp interface{}
		if temp == nil { // first
			if tp, ok := jsoncfg.jsonObject[keys[i]]; ok {
				_temp = tp
			} else {
				_temp = make(map[string]interface{})
				jsoncfg.jsonObject[keys[i]] = _temp
			}
		} else { //
			if tp, ok := temp.(map[string]interface{})[keys[i]]; ok {
				_temp = tp
			} else {
				_temp = make(map[string]interface{})
				temp.(map[string]interface{})[keys[i]] = _temp
			}
		}

		// find
		if _temp != nil {
			temp = _temp
		}
	}
	return nil
}

// readFileAsJSON 读取Json文件
func readFileAsJSON(path string, v interface{}) error {
	if len(path) == 0 {
		return pathNotExist("ReadFileAsJSON", path)
	}
	fp, err := os.OpenFile(path, os.O_RDONLY, 0)
	defer func() {
		if nil != fp {
			fp.Close()
		}
	}()

	if err == nil {
		st, stErr := fp.Stat()
		if stErr == nil {
			data := make([]byte, st.Size())
			_, err = fp.Read(data)
			if err == nil {
				return json.Unmarshal(data, v)
			}
		} else {
			err = stErr
		}
	}
	return err
}

// writeFileAsJSON 写入Json文件
func writeFileAsJSON(path string, v interface{}) error {
	if len(path) == 0 {
		return pathNotExist("WriteFileAsJSON", path)
	}
	fp, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	defer func() {
		if nil != fp {
			fp.Close()
		}
	}()

	if err == nil {
		data, err := json.Marshal(v)
		if err == nil {
			_, err := fp.Write(data)
			return err
		}
		return err
	}
	return err
}

// isFile 是否是文件
func isFile(path string) bool {
	_stat, _err := os.Stat(path)
	if _err == nil {
		return !_stat.IsDir()
	}
	return false
}

// isDir 是否是文件夹
func isDir(path string) bool {
	_stat, _err := os.Stat(path)
	if _err == nil {
		return _stat.IsDir()
	}
	return false
}

// getParentPath 范围最后一个'/'前的文字
func getParentPath(path string) string {
	if strings.Index(path, "\\") > -1 {
		return path[:strings.LastIndex(path, "\\")]
	}
	return path[:strings.LastIndex(path, "/")]
}

// pathNotExist 路径不存在的错误
func pathNotExist(op, path string) error {
	return &os.PathError{
		Op:   op,
		Path: path,
		Err:  os.ErrNotExist,
	}
}
