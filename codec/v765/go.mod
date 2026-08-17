module github.com/lrnxzz/graft/codec/v765

go 1.26.3

require (
	github.com/lrnxzz/graft v0.0.0-20260715002557-52d1baa1c5b9
	github.com/lrnxzz/graft/mojang v0.0.0-20260715002557-52d1baa1c5b9
	github.com/lrnxzz/graft/nbt v0.0.0-20260715002557-52d1baa1c5b9
)

require (
	github.com/andybalholm/brotli v1.2.1 // indirect
	github.com/klauspost/compress v1.18.6 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.72.0 // indirect
	golang.org/x/exp v0.0.0-20260709172345-9ea1abe57597 // indirect
	golang.org/x/sync v0.22.0 // indirect
)

replace github.com/lrnxzz/graft => ../..

replace github.com/lrnxzz/graft/mojang => ../../mojang

replace github.com/lrnxzz/graft/nbt => ../../nbt
