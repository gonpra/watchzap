build_w64:
	GOOS=windows CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc go build -o watchzap-x64.exe
	zip -r watchzap-windows-x64.zip watchzap-x64.exe
	sha256sum watchzap-windows-x64.zip >> CHECKSUM
build_l64:
	GOOS=linux go build -o watchzap-x64
	tar -cf watchzap-linux-x64.tar.gz watchzap-x64
	sha256sum watchzap-linux-x64.tar.gz >> CHECKSUM
clean:
	rm -rf watchzap-linux-x64.tar.gz watchzap-x64.exe watchzap-x64 watchzap-windows-x64.zip CHECKSUM
