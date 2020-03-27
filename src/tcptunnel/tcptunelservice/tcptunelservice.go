// Copyright (C) 2020 WuPeng <wupeng364@outlook.com>.
// Use of this source code is governed by an MIT-style.
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software,
// and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// 管道服务端

package main

import (
	"flag"
	"fmt"
	"net"
	"tcptunnel/tcpmsgexchanger"
	"tcptunnel/tcptunnelmanager"
)

func main() {
	// 获取需要加载的配置名字
	listenaddr := flag.String("listen", "0.0.0.0:8080", "web service listen addr")
	trunneladdr := flag.String("tunel", "0.0.0.0:8101", "tunel service addr")
	flag.Parse()

	// 服务地址
	fmt.Println("本地监听地址:", *listenaddr)
	fmt.Println("隧道监听地址:", *trunneladdr)
	taddr, err := net.ResolveTCPAddr("tcp4", *trunneladdr)
	if nil != err {
		panic(err)
	}
	laddr, err := net.ResolveTCPAddr("tcp4", *listenaddr)
	if nil != err {
		panic(err)
	}
	// 隧道服务启动
	TCPTunnelService := &tcptunnelmanager.TCPTunnelService{
		ServiceAddr: taddr,
	}
	go func() {
		err := TCPTunnelService.DoStart()
		if nil != err {
			panic(err)
		}
	}()
	err = doStartService(laddr, TCPTunnelService)
	if nil != err {
		panic(err)
	}
	fmt.Print("Ctrl+C退出程序")
	var sc string
	fmt.Scan(&sc)
	fmt.Println(sc)
}

// doStartService 启动服务端口
func doStartService(addr *net.TCPAddr, TCPTunnelService *tcptunnelmanager.TCPTunnelService) (err error) {
	listener, err := net.ListenTCP("tcp", addr)
	if nil == err {
		for {
			// 监听请求
			srcConn, err := listener.Accept()
			if nil != err {
				fmt.Println(err)
				continue
			}
			go (func() {
				destConn := TCPTunnelService.GetConn()
				if nil != destConn {
					defer (func() {
						if nil != srcConn {
							srcConn.Close()
						}
						TCPTunnelService.RelaseConn(destConn)
					})()
					// 交换数据
					TCPExchanger := &tcpmsgexchanger.TCPExchanger4HHTTP{}
					TCPExchanger.SetDebug(true)
					err = TCPExchanger.ExchangeData(srcConn, destConn)
					if nil != err {
						fmt.Println("交换数据错误: ", err)
					}
				} else {
					srcConn.Close()
				}
			})()
		}
	}
	return err
}
