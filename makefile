main: ptz.64 ptz.32

ptz.64: print-to-zebra.go
	go generate
	env GOOS=windows GOARCH=amd64 go build -buildmode=exe -o print-to-zebra-64bit.exe

ptz.32: print-to-zebra.go
	go generate
	env GOOS=windows GOARCH=386 go build -buildmode=exe -o print-to-zebra-32.exe


