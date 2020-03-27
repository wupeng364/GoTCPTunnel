// Copyright (C) 2019 WuPeng <wupeng364@outlook.com>.
// Use of this source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// 字符串处理工具

package strtool

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"io"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// ReplaceAll -> strings.Replace
func ReplaceAll(s, old, new string) string {
	return strings.Replace(s, old, new, -1)
}

// Parse2UnixPath 删除路径后面 /, 把\转换为/
func Parse2UnixPath(str string) string {
	if len(str) == 0 {
		return ""
	}
	return path.Clean(strings.Replace(str, "\\", "/", -1))
}

// GetPathParent 截取最后一个'/'前的文字
func GetPathParent(path string) string {
	if strings.Index(path, "\\") > -1 {
		return path[:strings.LastIndex(path, "\\")]
	}
	return path[:strings.LastIndex(path, "/")]
}

// GetPathName 截取最后一个'/'后的文字
func GetPathName(path string) string {
	if strings.Index(path, "\\") > -1 {
		return path[strings.LastIndex(path, "\\")+1:]
	}
	return path[strings.LastIndex(path, "/")+1:]
}

// ReadAsString 从io.Reader读取文字
func ReadAsString(src io.Reader) string {
	if nil == src {
		return ""
	}
	buf := make([]byte, 0)
	for {
		buftemp := make([]byte, 1024)
		nr, er := src.Read(buftemp)
		if nr > 0 {
			buf = append(buf, buftemp[:nr]...)
		}
		if er != nil {
			if er != io.EOF {
				return ""
			}
			break
		}
	}
	return string(buf)
}

// String2Bool 字符转bool
// true -> [1, t, T, true, TRUE, True]
// false -> [0, f, F, false, FALSE, False]
func String2Bool(str string) bool {
	switch strings.ToLower(str) {
	case "1", "t", "true":
		return true
	case "0", "f", "false":
		return false
	}
	return false
}

// Bool2String bool类型转string
func Bool2String(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// GetMD5 字符转MD5
func GetMD5(str string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(str))
	return hex.EncodeToString(md5Ctx.Sum(nil))
}

// GetMachineID 放回机器唯一标识符
// 计算MD5( 主机名 + 进程ID + 随机数 )
func GetMachineID() (string, error) {
	// 主机名
	host, err := os.Hostname()
	if nil != err {
		return "", err
	}
	// 进程ID
	pidstr := strconv.FormatInt(int64(os.Getpid()), 10)
	// 随机数
	uintByte := make([]byte, 4, 4)
	binary.BigEndian.PutUint32(uintByte, uint32(rand.Int31()))
	randhex := hex.EncodeToString(uintByte)
	// 计算MD5
	machineid := GetMD5(strings.Join([]string{host, pidstr, randhex}, ","))
	return machineid, nil
}

// GetRandom 生成随机字符
func GetRandom(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
