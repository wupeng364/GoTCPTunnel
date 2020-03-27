// Copyright (C) 2019 WuPeng <wupeng364@outlook.com>.
// Use of jsoncfg source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of jsoncfg software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and jsoncfg permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// 拓展对象-interface{}转各种类型

package types

// Object interface类型转换
type Object struct {
	// O 实例化时保存的原对象或指针
	O interface{}
}

// NewObject 新建一个object对象
func NewObject(obj interface{}) Object {
	return Object{O: obj}
}

// ToBool 转换为bool
func (obj Object) ToBool(d bool) bool {
	if nil == obj.O {
		return d
	}
	r, ok := obj.O.(bool)
	if ok {
		return r
	}
	return d
}

// ToString 转换为string
func (obj Object) ToString(d string) string {
	if nil == obj.O {
		return d
	}
	r, ok := obj.O.(string)
	if ok {
		return r
	}
	return d
}

// ToInt 转换为int
func (obj Object) ToInt(d int) int {
	if nil == obj.O {
		return d
	}
	r, ok := obj.O.(int)
	if ok {
		return r
	}
	return d
}

// ToInt32 转换为int32
func (obj Object) ToInt32(d int32) int32 {
	if nil == obj.O {
		return d
	}
	r, ok := obj.O.(int32)
	if ok {
		return r
	}
	return d
}

// ToInt64 转换为int64
func (obj Object) ToInt64(d int64) int64 {
	if nil == obj.O {
		return d
	}
	r, ok := obj.O.(int64)
	if ok {
		return r
	}
	return d
}

// ToFloat32 转换为float32
func (obj Object) ToFloat32(d float32) float32 {
	if nil == obj.O {
		return d
	}
	r, ok := obj.O.(float32)
	if ok {
		return r
	}
	return d
}

// ToFloat64 转换为Float64
func (obj Object) ToFloat64(d float64) float64 {
	if nil == obj.O {
		return d
	}
	r, ok := obj.O.(float64)
	if ok {
		return r
	}
	return d
}

// ToStrMap 转换为map[string]interface{}
func (obj Object) ToStrMap(d map[string]interface{}) map[string]interface{} {
	if nil == obj.O {
		return d
	}
	r, ok := obj.O.(map[string]interface{})
	if ok {
		return r
	}
	return d
}

// ToIntMap 转换为map[int]interface{}
func (obj Object) ToIntMap(d map[int]interface{}) map[int]interface{} {
	if nil == obj.O {
		return d
	}
	r, ok := obj.O.(map[int]interface{})
	if ok {
		return r
	}
	return d
}

// ToInt32Map 转换为map[int32]interface{}
func (obj Object) ToInt32Map(d map[int32]interface{}) map[int32]interface{} {
	if nil == obj.O {
		return d
	}
	r, ok := obj.O.(map[int32]interface{})
	if ok {
		return r
	}
	return d
}

// ToInt64Map 转换为map[int64]interface{}
func (obj Object) ToInt64Map(d map[int64]interface{}) map[int64]interface{} {
	if nil == obj.O {
		return d
	}
	r, ok := obj.O.(map[int64]interface{})
	if ok {
		return r
	}
	return d
}

// ToFloat32Map 转换为map[float32]interface{}
func (obj Object) ToFloat32Map(d map[float32]interface{}) map[float32]interface{} {
	if nil == obj.O {
		return d
	}
	r, ok := obj.O.(map[float32]interface{})
	if ok {
		return r
	}
	return d
}

// ToFloat64Map 转换为map[float64]interface{}
func (obj Object) ToFloat64Map(d map[float64]interface{}) map[float64]interface{} {
	if nil == obj.O {
		return d
	}
	r, ok := obj.O.(map[float64]interface{})
	if ok {
		return r
	}
	return d
}
