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
	"time"
)

const (
	// CMDMAXLEN 管理命令最大字符数
	CMDMAXLEN = 1024
	// CMDCONNECTCTRL 管理线程链接
	CMDCONNECTCTRL = "\r- doconnectctrl -\n"
	// CMDCONNECT 创建连接
	CMDCONNECT = "\r- doconnect -\n"
	// CMDCOUNTCONN 统计连接数
	CMDCOUNTCONN = "\r- countconn -\n"
	// CMDCLEARCONN 清理连接池
	CMDCLEARCONN = "\r- clearconn -\n"
	// CMDTRANSPORTSTART 开始传输
	CMDTRANSPORTSTART = "\r- transportstart -\n"
	// CMDCONNHEART 心跳包
	CMDCONNHEART = "\r- connheart -\n"
	// CMDOK 准备就绪
	CMDOK = "\r- ok -\n"
	// CMDRESET 重置链接
	CMDRESET = "\r- reset -\n"

	// CMDWTIMEOUT TCP写入超时
	CMDWTIMEOUT = time.Second * 30
	// CMDRTIMEOUT TCP读取超时
	CMDRTIMEOUT = time.Second * 60
)
