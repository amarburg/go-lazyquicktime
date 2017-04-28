

test:
	go test

profile:
	go test -cpuprofile cpu.prof -memprofile mem.prof -run TestExtractFrameLocalFileSource
	go tool pprof go-lazyquicktime.test cpu.prof

bench:
	go test -cpuprofile cpu.prof -memprofile mem.prof -bench=.

PHONY: test bench
