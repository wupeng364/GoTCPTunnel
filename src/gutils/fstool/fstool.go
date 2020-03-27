// Copyright (C) 2019 WuPeng <wupeng364@outlook.com>.
// Use of this source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// 文件工具

package fstool

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CopyCallback 文件复制回调(源路径, 目标路径, 错误信息) 返回的错误信息
type CopyCallback func(srcPath, dstPath string, err error) error

// MoveCallback 文件移动回调(源路径, 目标路径, 错误信息) 返回的错误信息
type MoveCallback func(srcPath, dstPath string, err error) error

// GetFileInfo 获取文件信息对象
func GetFileInfo(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// OpenFile 获取文件信息对象
func OpenFile(path string) (*os.File, error) {
	return os.Open(path)
}

// GetWriter 获取只写文件对象
// O_RDONLY: 只读模式(read-only)
// O_WRONLY: 只写模式(write-only)
// O_RDWR: 读写模式(read-write)
// O_APPEND: 追加模式(append)
// O_CREATE: 文件不存在就创建(create a new file if none exists.)
// O_EXCL: 与 O_CREATE 一起用, 构成一个新建文件的功能, 它要求文件必须不存在
// O_SYNC: 同步方式打开，即不使用缓存，直接写入硬盘
// O_TRUNC: 打开并清空文件
func GetWriter(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
}

// IsExist 文件/夹是否存在
func IsExist(path string) bool {
	_, err := GetFileInfo(path)
	return err == nil
}

// IsDir 是否是文件夹
func IsDir(path string) bool {
	stat, err := GetFileInfo(path)
	if err == nil {
		return stat.IsDir()
	}
	return false
}

// MkdirAll 创建文件夹-多级
func MkdirAll(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// Mkdir 创建文件夹
func Mkdir(path string) error {
	return os.Mkdir(path, os.ModePerm)
}

// IsFile 是否是文件
func IsFile(path string) bool {
	stat, err := GetFileInfo(path)
	if err == nil {
		return !stat.IsDir()
	}
	return false
}

// GetDirList 获取一级子目录名字(包含文件|文件夹,无序)
func GetDirList(path string) ([]string, error) {
	f, err := OpenFile(path)
	defer f.Close()
	if err != nil {
		return nil, err
	}

	list, err := f.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// GetFileSize 获取文件大小
func GetFileSize(path string) (int64, error) {
	f, err := OpenFile(path)
	if err != nil {
		return 0, err
	}
	fInfo, errStat := f.Stat()
	f.Close()
	if errStat != nil {
		return 0, errStat
	}
	return fInfo.Size(), nil
}

// GetCreateTime 获取创建时间
func GetCreateTime(path string) (time.Time, error) {
	return GetModifyTime(path)
}

// GetModifyTime 获取修改时间
func GetModifyTime(path string) (time.Time, error) {
	f, err := OpenFile(path)
	if err != nil {
		return time.Time{}, err
	}
	fInfo, errStat := f.Stat()
	f.Close()
	if errStat != nil {
		return time.Time{}, errStat
	}
	return fInfo.ModTime(), nil
}

// RemoveFile 删除文件
func RemoveFile(file string) error {
	if !IsExist(file) {
		return PathNotExist("RemoveFile", file)
	}
	return os.Remove(file)
}

// RemoveAll 删除文件
func RemoveAll(file string) error {
	if !IsExist(file) {
		return PathNotExist("RemoveAll", file)
	}
	return os.RemoveAll(file)
}

// Rename 重命名
func Rename(old, newName string) error {
	_path := filepath.Clean(old)
	if strings.Index(_path, "\\") > -1 {
		return os.Rename(old, old[:strings.LastIndex(old, "\\")+1]+newName)
	}
	return os.Rename(old, old[:strings.LastIndex(old, "/")+1]+newName)
}

// MoveFilesAcrossDisk 移动文件|文件夹,可跨分区移动(源路径, 目标路径, 重复覆盖, 重复忽略, 操作回调) 操作结果
func MoveFilesAcrossDisk(src, dst string, replace, ignore bool, callback MoveCallback) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)
	if src == dst && len(src) > 0 {
		return nil
	}
	if !IsExist(src) {
		return PathNotExist("MoveFilesAcrossDisk", src)
	}
	// 尝试本分区移动
	return MoveFiles(src, dst, replace, ignore, func(srcPath, dstPath string, mverror error) error {
		// 尝试跨分区移动
		if IsAcrossDiskError(mverror) {
			return MoveFileByCopying(srcPath, dstPath, replace, ignore, callback)
		}
		return callback(srcPath, dstPath, mverror)
	})
}

// MoveFiles 移动文件|夹, 如果存在的话就列表后逐一移动 (源路径, 目标路径, 重复覆盖, 重复忽略, 操作回调) 操作结果
func MoveFiles(src, dst string, replace, ignore bool, callback MoveCallback) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)
	if src == dst && len(src) > 0 {
		return nil
	}
	if !IsExist(src) {
		return PathNotExist("MoveFiles", src)
	}
	if IsExist(dst) {
		srcIsfile := IsFile(src)
		dstIsfile := IsFile(dst)
		var err error
		// src&dst都是文件夹, 则考虑合并规则, 处理里面的文件
		if !srcIsfile && !dstIsfile {
			list, _ := GetDirList(src)
			for _, val := range list {
				err = MoveFiles(src+"/"+val, dst+"/"+val, replace, ignore, callback)
				if err != nil {
					return err // 返回终止信号
				}
			}
			if IsExist(src) {
				return callback(src, dst, os.RemoveAll(src))
			}
			return nil
		}
		// 非合并模式只考虑删除目标位置
		if ignore {
			// 如果忽略则不返回错误
			return callback(src, dst, err)
		} else if replace {
			// 如果覆盖, 则先删除目标位置
			if dstIsfile {
				err = os.Remove(dst)
			} else {
				err = os.RemoveAll(dst)
			}
		} else {
			err = PathExist("MoveFiles", dst)
		}
		if nil != err {
			return callback(src, dst, err)
		}
		return callback(src, dst, os.Rename(src, dst))
	}
	// 如果目标文件/夹不存在就直接移动
	return callback(src, dst, os.Rename(src, dst))
}

// MoveFileByCopying 移动文件夹 - 跨分区-拷贝
func MoveFileByCopying(src, dst string, replace, ignore bool, callback MoveCallback) error {
	if src == dst && len(src) > 0 {
		return nil
	}
	if IsFile(src) {
		var err1 error
		err := CopyFile(src, dst, replace, ignore)
		if err != nil {
			err1 = callback(src, dst, err)
		}
		if err1 != nil {
			return err1
		} else if err != nil {
			return nil
		}
		return os.Remove(src)
	}
	// 复制文件夹
	err := CopyFiles(src, dst, replace, ignore, func(srcPath, dstPath string, err error) error {
		if err == nil && IsFile(srcPath) {
			err = os.Remove(srcPath)
		}
		return callback(srcPath, dstPath, err)
	})
	// 最后的清理
	if err == nil {
		err = os.RemoveAll(src)
	}
	return err
}

// CopyFile 复制文件(源路径, 目标路径, 重复覆盖, 重复忽略)
func CopyFile(src, dst string, replace, ignore bool) error {
	if src == dst && len(src) > 0 {
		return nil
	}
	if !IsFile(src) {
		return PathNotExist("CopyFile", src)
	}
	if IsExist(dst) {
		if replace {
			var err error
			if IsFile(src) {
				err = os.Remove(dst)
			} else {
				err = os.RemoveAll(dst)
			}
			if err != nil {
				return err
			}
		} else if ignore {
			return nil
		} else {
			return PathExist("CopyFile", dst)
		}
	}
	RSrc, err := OpenFile(src)
	defer func() {
		if nil != RSrc {
			RSrc.Close()
		}
	}()
	if err != nil {
		return err
	}
	WDst, err := GetWriter(dst)
	defer func() {
		if nil != WDst {
			WDst.Close()
		}
	}()
	if err != nil {
		return err
	}
	_, err = io.Copy(WDst, RSrc)
	return err
}

// CopyFiles 复制文件夹 (源路径, 目标路径, 重复覆盖, 重复忽略, 操作回调) 返回错误即可终止后续拷贝
// 如果callback返回错误非空, 该文件则为处理失败, 终止其他操作
// 如果callback返回nil则继续往下拷贝, 无论是否真的出错
func CopyFiles(src, dst string, replace, ignore bool, callback CopyCallback) error {
	if src == dst && len(src) > 0 {
		return nil
	}
	if !IsExist(src) {
		return callback(src, dst, PathNotExist("CopyFiles", src))
	}
	if IsFile(dst) {
		return callback(src, dst, PathExist("CopyFiles", dst))
	}
	dst = filepath.Clean(dst)
	src = filepath.Clean(src)
	srcLen := len(src)
	return filepath.Walk(src, func(s string, f os.FileInfo, err error) error {
		d := dst + s[srcLen:]
		if err == nil {
			if f.IsDir() {
				if !IsDir(d) {
					err = os.Mkdir(d, os.ModePerm)
				}
			} else {
				err = CopyFile(s, d, replace, ignore)
			}
		}
		return callback(s, d, err)
	})
}

// ReadFileAsJSON 读取Json文件
func ReadFileAsJSON(path string, v interface{}) error {
	if len(path) == 0 {
		return PathNotExist("ReadFileAsJSON", path)
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

// WriteFileAsJSON 写入Json文件
func WriteFileAsJSON(path string, v interface{}) error {
	if len(path) == 0 {
		return PathNotExist("WriteFileAsJSON", path)
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

// WriteTextFile 写入文本文件
func WriteTextFile(path, text string) error {
	if len(path) == 0 {
		return PathNotExist("WriteFile", path)
	}
	fp, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	defer func() {
		if nil != fp {
			fp.Close()
		}
	}()

	if err == nil {
		if err == nil {
			_, err := fp.Write([]byte(text))
			return err
		}
		return err
	}
	return err
}

// PathExist 路径已经存在的错误
func PathExist(op, path string) error {
	return &os.PathError{
		Op:   op,
		Path: path,
		Err:  os.ErrExist,
	}
}

// PathNotExist 路径不存在的错误
func PathNotExist(op, path string) error {
	return &os.PathError{
		Op:   op,
		Path: path,
		Err:  os.ErrNotExist,
	}
}

// IsExistError 是否是目标位置已经存在的错误
func IsExistError(err error) bool {
	if nil == err {
		return false
	}
	var cpErr error
	switch err := err.(type) {
	case *os.PathError:
		cpErr = err.Err
	}
	if cpErr == nil {
		return false
	}
	return cpErr.Error() == os.ErrExist.Error()
}

// IsNotExistError 是否是目标位置不存在的错误
func IsNotExistError(err error) bool {
	var cpErr error
	switch err := err.(type) {
	case *os.PathError:
		cpErr = err.Err
	}
	if cpErr == nil {
		return false
	}
	return cpErr.Error() == os.ErrNotExist.Error()
}

// IsAcrossDiskError 是否是跨磁盘错误
func IsAcrossDiskError(err error) bool {
	var cpErr error
	switch err := err.(type) {
	case *os.LinkError:
		cpErr = err.Err
	}
	if cpErr == nil {
		return false
	}
	return cpErr.Error() == "The system cannot move the file to a different disk drive."
}
