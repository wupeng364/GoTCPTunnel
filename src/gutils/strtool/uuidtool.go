// Copyright (C) 2019 WuPeng <wupeng364@outlook.com>.
// Use of this source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// UUID工具

package strtool

import (
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"time"
)

// 初始化机器ID, 避免每次都算
var machineByte []byte

// 初始化机器ID, 避免每次都算
func init() {
	machineid, err := GetMachineID()
	if nil != err {
		panic(err)
	}
	machineByte, err = hex.DecodeString(machineid)
	if nil != err {
		panic(err)
	}
}

// GetUUID 获取唯一ID
func GetUUID() string {
	// 时间
	uintByte64 := make([]byte, 8, 8)
	binary.BigEndian.PutUint64(uintByte64, uint64(time.Now().UnixNano()))
	// 随机数
	uintByte32 := make([]byte, 4, 4)
	binary.BigEndian.PutUint32(uintByte32, uint32(rand.Int31()))
	// baseID
	baseID := make([]byte, 0, 16)
	baseID = append(baseID, machineByte[0:4]...) // 4
	baseID = append(baseID, uintByte64...)       // 8
	baseID = append(baseID, uintByte32...)       // 4

	gid := make([]byte, 0, 36)
	id := []byte(hex.EncodeToString(baseID))
	gid = append(gid, id[0:8]...)
	gid = append(gid, '-')
	gid = append(gid, id[8:12]...)
	gid = append(gid, '-')
	gid = append(gid, id[12:16]...)
	gid = append(gid, '-')
	gid = append(gid, id[16:20]...)
	gid = append(gid, '-')
	gid = append(gid, id[20:]...)
	return string(gid)
}
