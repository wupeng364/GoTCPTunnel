// Copyright (C) 2020 WuPeng <wupeng364@outlook.com>.
// Use of this source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// 管道客户端

package main

import (
	"flag"
	"fmt"
	"net"
	"tcptunnel/tcpmsgexchanger"
	"tcptunnel/tcptunnelmanager"
	"time"
)

func main() {
	// 获取需要加载的配置名字
	serveraddr := flag.String("server", "127.0.0.1:8101", "tunnel server addr")
	proxyaddr := flag.String("proxy", "192.168.2.8:80", "proxy server addr")
	flag.Parse()

	// 服务地址
	fmt.Println("隧道服务地址:", *serveraddr)
	fmt.Println("远程代理地址:", *proxyaddr)
	serviceAddr, err := net.ResolveTCPAddr("tcp4", *serveraddr)
	if nil != err {
		panic(err)
	}
	// 连接管理服务
	TCPTunnelClient := &tcptunnelmanager.TCPTunnelConnector{
		ServiceAddr: serviceAddr,
	}
	// 当收到链接后执行
	TCPTunnelClient.SetTransportCallback(func(remote *net.TCPConn, relase func()) (err error) {
		defer (func() {
			relase()
		})()
		// 等待连接响应, 阻塞执行
		_, err = remote.Read(make([]byte, 0))
		if nil == err {
			// 连接代理目标服务器
			destAddr, err := net.ResolveTCPAddr("tcp4", *proxyaddr)
			if nil == err {
				destConn, err := net.DialTCP("tcp4", nil, destAddr)
				defer (func() {
					if nil != destConn {
						destConn.Close()
					}
				})()
				if nil == err {
					// TCP消息交换
					TCPExchanger := &tcpmsgexchanger.TCPExchanger4HHTTP{}
					TCPExchanger.SetDebug(true)
					err = TCPExchanger.ExchangeData(remote, destConn)
				}
			}
		}
		if nil != err {
			fmt.Println("转发数据出现错误: ", err)
		}
		return err
	})
	for {
		err := TCPTunnelClient.DoConnect()
		if nil != err {
			fmt.Println(err)
		}
		fmt.Println("隧道连接异常,正在重连")
		time.Sleep(time.Duration(1) * time.Second)
	}

	// fmt.Print("Ctrl+C退出程序")
	// var sc string
	// fmt.Scan(&sc)
	// fmt.Println(sc)
}
