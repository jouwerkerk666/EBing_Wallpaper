all:	build install

build:
	go build -o EBing_Wallpaper EBing_Wallpaper.go
	strip ./EBing_Wallpaper
install:
	sudo cp ./EBing_Wallpaper /usr/local/bin/EBing_Wallpaper
