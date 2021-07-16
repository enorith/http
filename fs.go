package http

import (
	"fmt"
	"sync"

	"github.com/valyala/fasthttp"
)

var (
	fsHandlers = make(map[string]fasthttp.RequestHandler)
	mu         sync.RWMutex
)

func GetFsHandler(root string, stripSlashes int) fasthttp.RequestHandler {
	key := fmt.Sprintf("%s%d", root, stripSlashes)
	mu.RLock()
	h, ok := fsHandlers[key]
	mu.RUnlock()
	if ok {
		return h
	}
	mu.Lock()
	defer mu.Unlock()
	h = fasthttp.FSHandler(root, stripSlashes)
	fsHandlers[key] = h
	return h
}
