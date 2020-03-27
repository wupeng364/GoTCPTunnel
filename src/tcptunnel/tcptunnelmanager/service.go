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
	"errors"
	"fmt"
	"gutils/strtool"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

// TCPTunnelService TCP隧道服务端
type TCPTunnelService struct {
	ServiceAddr *net.TCPAddr            // 管道服务端口
	conns       map[string]*net.TCPConn // 连上来的线程
	ctlConn     *net.TCPConn            // 控制线程, 只能连接一次
	isDebug     bool                    // 是否输出调试信息
	serviceID   string                  // 实例ID
	lock        *sync.RWMutex
}

// printInfo 打印信息
func (service *TCPTunnelService) printInfo(a ...interface{}) {
	if service.isDebug {
		fmt.Println("["+service.serviceID+"]", a)
	}
}

// GetID 获取实例ID
func (service *TCPTunnelService) GetID() string {
	return service.serviceID
}

// DoStart 启动隧道服务
func (service *TCPTunnelService) DoStart() (err error) {
	service.lock = new(sync.RWMutex)
	if len(service.serviceID) == 0 {
		service.serviceID = strtool.GetUUID()
	}
	service.sendConnHeart() // 启动心跳检测
	listener, err := net.ListenTCP("tcp", service.ServiceAddr)
	if nil == err {
		for {
			conn, err := listener.AcceptTCP()
			if nil != err {
				service.printInfo("AcceptTCP error: ", err)
				continue
			}
			// 1. 检查是否是控制线程连接
			cmd := service.getCMD(conn)
			if nil == service.ctlConn && CMDCONNECTCTRL != cmd {
				continue
			}
			switch cmd {
			case CMDCONNECTCTRL: // 这是管理线程链接, 监听着, 不断开
				// 记录链接, 并清空之前的连接
				service.ctlConn = conn
				service.clearConn()
				go service.doConnCtrlAdapter()
				break
			case CMDCONNECT: // 客户端新建链接请求
				service.lock.Lock()
				service.conns[conn.RemoteAddr().String()] = conn
				service.lock.Unlock()
				break
			default:
				conn.Close()
				break
			}
		}
	}
	return err
}

// sendConnHeart 保持心跳
func (service *TCPTunnelService) sendConnHeart() {
	go (func() {
		for {
			service.lock.RLock()
			if len(service.conns) > 0 {
				for key, val := range service.conns {
					go func(key string, val *net.TCPConn) {
						service.printInfo("sendConnHeart: ", key)
						_, err := val.Write([]byte(CMDCONNHEART))
						if nil == err {
							cmd := service.getCMD(val)
							if cmd != CMDOK {
								err = errors.New("Connect heart response is error, responsed: " + cmd)
							}
						}
						if nil != err {
							val.Close()
							service.lock.Lock()
							defer service.lock.Unlock()
							delete(service.conns, key)
							service.printInfo("deleteConn: ", key, err)
						}
					}(key, val)
				}
			}
			service.lock.RUnlock()
			time.Sleep(time.Duration(5) * time.Second)
		}
	})()
}

// doConnCtrlAdapter 启动控制侦听
func (service *TCPTunnelService) doConnCtrlAdapter() {
	if nil == service.ctlConn {
		return
	}
	defer (func() {
		service.clearConn()
		service.ctlConn.Close()
		service.ctlConn = nil
	})()
	for {
		cmd := service.getCMD(service.ctlConn)
		service.printInfo("CMD:", cmd)
		if len(cmd) > 0 {
			var err error
			switch cmd {
			case CMDCOUNTCONN:
				_, err = service.ctlConn.Write([]byte(strconv.Itoa(len(service.conns))))
				break
			default:
				_, err = service.ctlConn.Write([]byte("401: cmd not support!"))
				break
			}
			if nil != err {
				fmt.Println("隧道终端-控制器连接异常,正在断开链接", "指令回复异常")
				break
			}
		} else {
			fmt.Println("隧道终端-控制器连接异常,正在断开链接", "无法读取到指令")
			break
		}
	}
}

// GetConn 获取一个空闲连接, 可用链接-1
func (service *TCPTunnelService) GetConn() *net.TCPConn {
	service.lock.Lock()
	defer service.lock.Unlock()
	if len(service.conns) > 0 {
		for key, conn := range service.conns {
			delete(service.conns, key)
			_, err := conn.Write([]byte(CMDTRANSPORTSTART))
			if nil != err {
				service.printInfo("send transport start cmd error: ", err)
				continue
			}
			cmd := service.getCMD(conn)
			if len(cmd) == 0 {
				continue
			}
			return conn
		}
	}
	return nil
}

// RelaseConn 释放连接, 如不释放, 隧道终端可能会一直创建新的链接
func (service *TCPTunnelService) RelaseConn(conn *net.TCPConn) {
	cmd := service.getCMD(conn)
	if cmd == CMDRESET {
		service.lock.Lock()
		defer service.lock.Unlock()
		service.conns[conn.RemoteAddr().String()] = conn
		service.printInfo("relaseConn", conn.RemoteAddr().String())
	}
}

// getCMD 读取隧道响应消息
func (service *TCPTunnelService) getCMD(conn *net.TCPConn) string {
	b := make([]byte, CMDMAXLEN)
	n, err := conn.Read(b)
	if nil == err || err == io.EOF {
		return string(b[:n])
	}
	return ""
}

// sendCmd 发送控制指令
func (service *TCPTunnelService) sendCMD(conn *net.TCPConn, cmd string) error {
	if nil == conn {
		return errors.New("conn is nil")
	}
	return nil
}

// clearConn 关闭所有连接
func (service *TCPTunnelService) clearConn() {
	service.lock.Lock()
	if len(service.conns) > 0 {
		for key, val := range service.conns {
			delete(service.conns, key)
			val.Close()
			service.printInfo("closeConn: ", key)
		}
	}
	service.conns = make(map[string]*net.TCPConn)
	service.lock.Unlock()
}
