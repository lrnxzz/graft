module github.com/lrnxzz/graft

go 1.26.3

require (
	github.com/lrnxzz/graft/codec/v765 v765.0.0-20260715043208-c635497e8677
	github.com/lrnxzz/graft/mojang v0.0.0-20260715002557-52d1baa1c5b9
	github.com/lrnxzz/graft/nbt v0.0.0-20260715002557-52d1baa1c5b9
	github.com/spf13/cobra v1.10.2
	golang.org/x/sync v0.22.0
)

require (
	github.com/andybalholm/brotli v1.2.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/compress v1.18.6 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.72.0 // indirect
)

replace github.com/lrnxzz/graft/codec/v765 => ./codec/v765

replace github.com/lrnxzz/graft/mojang => ./mojang

replace github.com/lrnxzz/graft/nbt => ./nbt
