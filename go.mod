module github.com/c-bata/vtermtest

go 1.19

require (
	github.com/creack/pty v1.1.24
	github.com/mattn/go-libvterm v0.0.0-20220218002314-74b0d3133396
	github.com/mattn/go-runewidth v0.0.16
)

require (
	github.com/mattn/go-pointer v0.0.1 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
)

replace github.com/mattn/go-libvterm => github.com/c-bata/go-libvterm v0.0.0-20250813102408-766a93136d87
