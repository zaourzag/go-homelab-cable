set CGO_CFLAGS=-IC:\libvlc\include
set CGO_LDFLAGS=-LC:\libvlc
go build -o glhc.exe cmds\cli\main.go