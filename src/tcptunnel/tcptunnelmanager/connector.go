// Copyright (C) 2020 WuPeng <wupeng364@outlook.com>.
// Use of this source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package tcptunnelmanager

import (
	"fmt"
	"gutils/strtool"
	"io"
	"net"
	"strconv"
	"time"
)

// onTransport 当链接上隧道后的回调函数, conn: 链接对象, release: 释放资源
type onTransport func(conn *net.TCPConn, release func()) error

// TCPTunnelConnector TCP隧道客户端
type TCPTunnelConnector struct {
	ServiceAddr  *net.TCPAddr
	OnTransport  onTransport
	MaxCount     int64  // 保持空闲连接数
	connectorID  string // 实例ID
	currentCount int64
	isDebug      bool // 是否输出调试信息
}

// SetTransportCallback 设置当链接上隧道后的回调函数
func (connector *TCPTunnelConnector) SetTransportCallback(fuc onTransport) {
	connector.OnTransport = fuc
}

// printInfo 打印信息
func (connector *TCPTunnelConnector) printInfo(a ...interface{}) {
	if connector.isDebug {
		fmt.Println("["+connector.connectorID+"]", a)
	}
}

// GetID 获取实例ID
func (connector *TCPTunnelConnector) GetID() string {
	return connector.connectorID
}

// getCMD 读取隧道响应消息
func (connector *TCPTunnelConnector) getCMD(conn net.Conn) string {
	b := make([]byte, CMDMAXLEN)
	n, err := conn.Read(b)
	if nil == err || err == io.EOF {
		return string(b[:n])
	}
	return ""
}

// DoConnect 连接隧道服务
func (connector *TCPTunnelConnector) DoConnect() (err error) {
	if connector.MaxCount == 0 {
		connector.MaxCount = 50
	}
	if len(connector.connectorID) == 0 {
		connector.connectorID = strtool.GetUUID()
	}
	conn, err := net.DialTCP("tcp4", nil, connector.ServiceAddr)
	defer (func() {
		if nil != conn {
			conn.Close()
		}
	})()
	if nil == err {
		// 说明连接上服务端了
		// 1. 先清空服务端现有隧道连接缓存
		_, err = conn.Write([]byte(CMDCONNECTCTRL))
		if nil == err {
			for {
				// 2. 查询服务端的连接情况
				_, err = conn.Write([]byte(CMDCOUNTCONN))
				if nil == err {
					connector.currentCount, err = strconv.ParseInt(connector.getCMD(conn), 10, 64)
				}
				if nil == err {
					// 3. 如果个数不够则需要创建新连接
					if connector.MaxCount > connector.currentCount {
						connector.doAddConnect()
					} else {
						time.Sleep(time.Duration(500) * time.Millisecond)
					}
				}
				if nil != err {
					return err
				}
			}
		}
	}
	return err
}

// doListen 监听是否是有数据发送过来
func (connector *TCPTunnelConnector) doListen(conn *net.TCPConn) {
	if nil != conn {
		for {
			cmd := connector.getCMD(conn)
			connector.printInfo("Listen-MSG: ", cmd)
			if cmd == CMDTRANSPORTSTART {
				_, err := conn.Write([]byte(CMDOK))
				if nil != err {
					conn.Close()
					break
				}
				if nil != connector.OnTransport {
					connector.OnTransport(conn, func() {
						_, err := conn.Write([]byte(CMDRESET))
						if nil != err {
							conn.Close()
						} else {
							connector.doListen(conn)
						}
					})
				}
			} else if cmd == CMDCONNHEART {
				_, err := conn.Write([]byte(CMDOK))
				if nil != err {
					conn.Close()
					break
				}
			} else {
				// 不识别的信号, 断开链接
				conn.Close()
				break
			}
		}
	}
}

// doAddConnect 添加隧道空闲连接
func (connector *TCPTunnelConnector) doAddConnect() error {
	conn, err := net.DialTCP("tcp4", nil, connector.ServiceAddr)
	if nil != err {
		return err
	}
	// 发送连接请求
	_, err = conn.Write([]byte(CMDCONNECT))
	if nil != err {
		conn.Close()
		return err
	}
	// 执行回调
	go connector.doListen(conn)
	return nil
}
