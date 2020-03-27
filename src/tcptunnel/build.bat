set B_HOME=%cd%
cd ../../
set GOPATH=%cd%

rem build windows
cd %B_HOME%/tcptunelclient
go install
cd %B_HOME%/tcptunelservice
go install

rem build linux
set GOOS=linux
set GOARCH=amd64
cd %B_HOME%/tcptunelclient
go install
cd %B_HOME%/tcptunelservice
go install