// Copyright (C) 2020 WuPeng <wupeng364@outlook.com>.
// Use of this source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// HTTP协议数据交换器

package tcpmsgexchanger

import (
	"fmt"
	"gutils/strtool"
	"io"
	"net"
	"strconv"
	"strings"
)

const (
	// HTTPBODYSPLITTER HTTP协议规定用回车和换行符分割数据
	HTTPBODYSPLITTER = "\r\n\r\n"
	// HTTPTRANSFERENCODINGENDCODE 分段传输时报文结束符号
	HTTPTRANSFERENCODINGENDCODE = "\r\n0\r\n"
	// HTTPHEADERCONTENTLENGTH HTTP头信息-作为报文是否结束的依据, 但有可能没有该字段
	HTTPHEADERCONTENTLENGTH = "Content-Length"
	// HTTPHEADERTRANSFERENCODING HTTP头信息-表示使用分段传输
	HTTPHEADERTRANSFERENCODING = "Transfer-Encoding"
	// HTTPHEADERMAXLENGTH HTTP头信息最大解析长度
	HTTPHEADERMAXLENGTH = 1024 * 1024 * 2
)

// TCPExchanger4HHTTP 检查HTTP报文信息
// 1. 是否是http报文, 2. 当前报文是否接收完成
type TCPExchanger4HHTTP struct {
	isDebug        bool              // 是否调试输出
	isExchange     bool              // 是否是双向交换数据
	exchengerID    string            // 处理id
	headerEndIndex int64             // 头信息结束位置
	bodyEndIndex   int64             // 内容结束位置
	headers        map[string]string // http头信息
	receivedLength int64             // 总计接收了多少数据
	receivedByte   []byte            // 临时缓存数据
}

// printInfo 打印信息
func (exchanger *TCPExchanger4HHTTP) printInfo(a ...interface{}) {
	if exchanger.isDebug {
		fmt.Println("["+exchanger.exchengerID+"]", a)
	}
}

// SetDebug 设置是否输出日志
func (exchanger *TCPExchanger4HHTTP) SetDebug(b bool) {
	exchanger.isDebug = b
}

// GetID 获取操作ID
func (exchanger *TCPExchanger4HHTTP) GetID() string {
	return exchanger.exchengerID
}

// ExchangeData 双向交换数据 SRC <-> DEST, 双向交换数据, 操作id不会变
func (exchanger *TCPExchanger4HHTTP) ExchangeData(src net.Conn, dest net.Conn) error {
	exchanger.isExchange = true
	exchanger.exchengerID = strtool.GetUUID()
	exchanger.printInfo("SRC --> DEST(" + src.RemoteAddr().String() + " ---> " + dest.RemoteAddr().String() + ")")
	err := exchanger.SendData(src, dest)
	if nil != err {
		return err
	}
	exchanger.printInfo("DEST --> SRC(" + dest.RemoteAddr().String() + " ---> " + src.RemoteAddr().String() + ")")
	return exchanger.SendData(dest, src)
}

// SendData 单向交换数据 SRC -> DEST, 单向交换数据, 操作id每次都不一样
func (exchanger *TCPExchanger4HHTTP) SendData(src net.Conn, dest net.Conn) error {
	if !exchanger.isExchange {
		exchanger.exchengerID = strtool.GetUUID()
		exchanger.printInfo("SRC --> DEST(" + src.RemoteAddr().String() + " ---> " + dest.RemoteAddr().String() + ")")
	}
	// 清空状态数据
	exchanger.headerEndIndex = int64(0)
	exchanger.bodyEndIndex = int64(0)
	exchanger.receivedLength = int64(0)
	exchanger.receivedByte = make([]byte, 0)
	exchanger.headers = make(map[string]string, 0)

	for {
		// 从SRC机器读取数据
		byteSrc := make([]byte, HTTPHEADERMAXLENGTH)
		nSrc, errSrc := src.Read(byteSrc)
		if nil != errSrc {
			if errSrc != io.EOF {
				exchanger.printInfo(errSrc)
				return errSrc
			}
			break
		}
		// 先发送回去一些数据
		_, errDest := dest.Write(byteSrc[:nSrc])
		if nil != errDest {
			exchanger.printInfo(errDest)
			return errSrc
		}
		// 检查器接受数据
		exchanger.receive(byteSrc[:nSrc])
		// 检查是否接受完毕
		if exchanger.isEnd() {
			break
		}
	}
	return nil
}

// receive 接受字节, 用于刷新状态
func (exchanger *TCPExchanger4HHTTP) receive(bt []byte) {
	// 累加接受的字节数
	receivedStr := string(bt)
	exchanger.printInfo("检查器接收到的长度:", exchanger.receivedLength, len(bt))
	exchanger.receivedLength = exchanger.receivedLength + int64(len(bt))
	//exchanger.printInfo( receivedStr )
	// 如果没有解析过头, 则需要解析头信息
	if exchanger.headerEndIndex <= 0 {
		exchanger.receivedByte = append(exchanger.receivedByte, bt[:len(bt)]...)
		receivedHeaderStr := string(exchanger.receivedByte)
		// 第一个换行符 - 请求行+头信息
		exchanger.headerEndIndex = int64(strings.Index(receivedHeaderStr, HTTPBODYSPLITTER))
		if exchanger.headerEndIndex > -1 {
			exchanger.printInfo("扫描Header信息!")
			// 保存头信息
			exchanger.headers = exchanger.str2Headers(receivedHeaderStr[:exchanger.headerEndIndex])
			exchanger.receivedByte = make([]byte, 0)
		} else {
			// 这个包有可能是HTTPBody, 或者其他协议, 如果一直接受下去内存可能会爆炸
			// 我们假设头信息不会超过2M, 如果超过这个长度报文缓存, 那我们则丢弃清空
			// 同时由于之前可能没有接受到头信息, 所以Body结束位置只能每次都判断计算
			if len(exchanger.receivedByte) >= HTTPHEADERMAXLENGTH {
				exchanger.receivedByte = make([]byte, 0)
				// > -1 说明是分段传输,检查到结束符号 0\r\n\r\n
				exchanger.bodyEndIndex = int64(strings.Index(receivedStr, HTTPTRANSFERENCODINGENDCODE))
				exchanger.printInfo("分段传输结束符检查-无头:", exchanger.bodyEndIndex)
				if exchanger.bodyEndIndex < 0 {
					// 普通结束符号
					exchanger.bodyEndIndex = int64(strings.LastIndex(receivedStr, HTTPBODYSPLITTER))
					exchanger.printInfo("标准结束符检查-无头:", exchanger.bodyEndIndex)
				}
				// 在同一个包内的位置
				if exchanger.bodyEndIndex > 0 && exchanger.receivedLength > int64(len(bt)) {
					exchanger.bodyEndIndex = exchanger.receivedLength + exchanger.bodyEndIndex
				}
			}

		}
	}
	// 如果头信息解析好了, 则需要扫描是否存在结束换行符
	if exchanger.headerEndIndex > 0 {
		// 检查是否有Transfer-Encoding: chunked, 说明是分段传输, 需要判断结束符号 0\r\n\r\n
		if val, exist := exchanger.headers[HTTPHEADERTRANSFERENCODING]; exist && strings.Replace(val, " ", "", -1) == "chunked" {
			// > -1 说明是分段传输,检查到结束符号 0\r\n\r\n
			exchanger.bodyEndIndex = int64(strings.Index(receivedStr, HTTPTRANSFERENCODINGENDCODE))
			exchanger.printInfo("分段传输结束符检查:", exchanger.bodyEndIndex)
		} else {
			// 普通结束符号
			exchanger.bodyEndIndex = int64(strings.LastIndex(receivedStr, HTTPBODYSPLITTER))
			exchanger.printInfo("标准结束符检查:", exchanger.bodyEndIndex)
		}
		// 在同一个包内的位置
		if exchanger.bodyEndIndex > 0 && exchanger.receivedLength > int64(len(bt)) {
			exchanger.bodyEndIndex = exchanger.receivedLength + exchanger.bodyEndIndex
		}
	}
}

// isEnd 消息体是否结束了
func (exchanger *TCPExchanger4HHTTP) isEnd() bool {
	// 是否接受到头信息
	if exchanger.headerEndIndex <= 0 {
		return false
	}
	// 1. 如果有头信息
	// 检查是否有 Content-Length 设置
	if val, exist := exchanger.headers[HTTPHEADERCONTENTLENGTH]; exist {
		// 获取一次, 内容长度
		length, err := strconv.ParseInt(strings.Replace(val, " ", "", -1), 10, 64)
		if nil != err {
			fmt.Println(err)
			return true
		}
		// 当前数据长度是否大于 Content-Length长度 + 头信息下标 + 头信息分隔符长度
		exchanger.printInfo("Content-Length长度检查:", exchanger.receivedLength, length, exchanger.headerEndIndex, len(HTTPBODYSPLITTER))
		if exchanger.receivedLength >= int64(length)+int64(exchanger.headerEndIndex)+int64(len(HTTPBODYSPLITTER)) {
			exchanger.printInfo("Content-Length长度检查通过!")
			return true
		}
		return false
	}
	// 2. 如果没有指定内容长度
	// 检查是否发现了结束符
	if exchanger.bodyEndIndex > 0 {
		exchanger.printInfo("Body结束符检查:", exchanger.bodyEndIndex, exchanger.headerEndIndex)
		// 如果开始和结束位置相等, 则需要判断这段数据是否一共就只有这么长
		if exchanger.bodyEndIndex == exchanger.headerEndIndex {
			if exchanger.receivedLength == int64(exchanger.bodyEndIndex)+int64(len(HTTPBODYSPLITTER)) {
				exchanger.printInfo("Body结束符检查通过!")
				return true // 数据就这么长了
			}

		} else if exchanger.bodyEndIndex > exchanger.headerEndIndex {
			exchanger.printInfo("Body结束符检查通过!")
			return true
		}
	}
	return false
}

// str2Headers 获取头信息
func (exchanger *TCPExchanger4HHTTP) str2Headers(body string) map[string]string {
	res := make(map[string]string, 0)
	if len(body) > 0 {
		// 每个头信息占一行
		lines := strings.Split(body, "\r\n")
		for i := 0; i < len(lines); i++ {
			// 判断是否是头信息行
			if strings.Index(lines[i], ":") > -1 {
				line := strings.Split(lines[i], ":")
				// 替换头信息空格 并  分割字符
				res[line[0]] = line[1]
			}
		}

	}
	return res
}
