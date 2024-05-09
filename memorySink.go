package slog

import (
	"bytes"
	"go.uber.org/zap"
	"net/url"
	"sync"
)

var (
	_ zap.Sink = &memorySink{}
)

var (
	sinkRegistry struct {
		once sync.Once
		sync.Mutex
		m map[string]*memorySink
	}
)

func initRegistry() {
	sinkRegistry.m = make(map[string]*memorySink)
	if err := registerMemorySinkFactory(); err != nil {
		panic(err)
	}
}

func registerMemorySinkFactory() error {
	return zap.RegisterSink("memory", func(u *url.URL) (zap.Sink, error) {
		if len(u.Host) == 0 {
			panic("invalid memory sink name")
		}

		sinkRegistry.Lock()
		sink := sinkRegistry.m[u.Host]
		sinkRegistry.Unlock()
		if sink == nil {
			panic("impossible")
		}

		return sink, nil
	})
}

type memorySink struct {
	buf bytes.Buffer
}

func (s *memorySink) Write(p []byte) (n int, err error) {
	return s.buf.Write(p)
}

func (s *memorySink) Sync() error {
	return nil
}

func (s *memorySink) Close() error {
	return nil
}
