package httpProxyPool

import (
	"errors"
	"fmt"
	"github.com/B9O2/Inspector/useful"
	. "github.com/Kumengda/httpProxyPool/runtime"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
)

type HttpProxyPool struct {
	httpClients  []*HttpProxyClientInfo
	timeout      int
	lock         sync.Mutex
	retryHandler func(err error, r *http.Request, response *http.Response, proxyUrl string) (bool, bool)
	maxRetryNum  int
	refresh      bool
}

func init() {
	Init()
}

func NewHttpProxyPool(timeout int) *HttpProxyPool {
	return &HttpProxyPool{
		timeout: timeout,
	}
}

func (h *HttpProxyPool) RegistryClientHandler(handler func(timeout int) []*HttpProxyClientInfo) {
	h.httpClients = handler(h.timeout)
	go func() {
		for {
			if h.refresh {
				MainInsp.Print(useful.DEBUG, useful.Text("client刷新指令收到,开始刷新...."))
				h.lock.Lock()
				h.httpClients = nil
				h.httpClients = handler(h.timeout)
				h.refresh = false
				h.lock.Unlock()
				MainInsp.Print(useful.DEBUG, useful.Text("client刷新结束"))
			}
		}
	}()
}

func (h *HttpProxyPool) getHttpClient() (uint32, error) {
	h.lock.Lock()
	sort.Slice(h.httpClients, func(i, j int) bool {
		return h.httpClients[i].UsingNum < h.httpClients[j].UsingNum
	})
	h.lock.Unlock()
	if len(h.httpClients) == 0 {
		return 0, errors.New("http client empty")
	}
	MainInsp.Print(useful.DEBUG, useful.Text(fmt.Sprintf("获取到client,当前client代理地址为%s使用量为%d", h.httpClients[0].ProxyUrl, h.httpClients[0].UsingNum)))
	return h.httpClients[0].hashCode, nil
}
func (h *HttpProxyPool) SetRetryHandler(handler func(err error, r *http.Request, response *http.Response, proxyUrl string) (bool, bool), maxRetryNum int) {
	h.retryHandler = handler
	h.maxRetryNum = maxRetryNum
}
func (h *HttpProxyPool) Do(doReq *http.Request) (*http.Response, error) {
	retry := 0
	for {
		if h.maxRetryNum == -1 {
			MainInsp.Print(useful.WARN, useful.Text("重试次数为无限重试"))
			retry = h.maxRetryNum - 1
		}
		req, resp, proxyUrl, err := h.doReq(doReq)
		if err != nil {
			if err.Error() == "http client empty" {
				return nil, err
			}
			if h.retryHandler == nil {
				return resp, err
			}
			if retry >= h.maxRetryNum {
				return resp, err
			}
			isContinue, isRefresh := h.retryHandler(err, req, resp, proxyUrl)
			if isRefresh {
				h.refresh = true
			}
			if isContinue {
				retry++
				continue
			} else {
				return resp, err
			}
		}
		return resp, err
	}
}
func (h *HttpProxyPool) doReq(req *http.Request) (*http.Request, *http.Response, string, error) {
	var httpProxyClientInfo *HttpProxyClientInfo
	clientHashCode, err := h.getHttpClient()
	if err != nil {
		return nil, nil, "", err
	}
	for _, v := range h.httpClients {
		if v.hashCode == clientHashCode {
			httpProxyClientInfo = v
		}
	}
	if httpProxyClientInfo == nil {
		return nil, nil, "", errors.New("http client empty")
	}
	atomic.AddInt32(&httpProxyClientInfo.UsingNum, 1)
	resp, err := httpProxyClientInfo.HttpClient.Do(req)
	atomic.AddInt32(&httpProxyClientInfo.UsingNum, -1)
	return req, resp, httpProxyClientInfo.ProxyUrl, err
}
