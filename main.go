package main

import (
	"fmt"
	"github.com/Kumengda/httpProxyPool/httpProxyPool"
	"github.com/Kumengda/httpProxyPool/testProxy"
	"net/http"
	"strings"
	"time"
)

func main() {
	req, _ := http.NewRequest("GET", "https://www.google.com", nil)
	aaa := httpProxyPool.NewHttpProxyPool(10)
	aaa.RegistryClientHandler(func(timeout int) []*httpProxyPool.HttpProxyClientInfo {
		var httpProxyPools []*httpProxyPool.HttpProxyClientInfo
		a, _ := testProxy.GetProxy("http://192.168.104.151:50001/collect/task/proxy-ip")
		for _, v := range a {
			info, err := httpProxyPool.NewHttpSockets5ClientInfo(v.Username, v.Password, v.Host, timeout)
			if err != nil {
				continue
			}
			httpProxyPools = append(httpProxyPools, &info)
		}
		return httpProxyPools
	})

	aaa.SetRetryHandler(func(err error, r *http.Request, response *http.Response, proxyUrl string) (bool, bool) {
		if err != nil {
			if strings.Contains(err.Error(), "Proxy Authentication Required") || strings.Contains(err.Error(), "unexpected EOF") {
				return false, true
			}
		}
		return false, false
	}, 10)

	for {
		go func() {
			do, err := aaa.Do(req)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(do.StatusCode)
		}()
		time.Sleep(time.Millisecond * 10)
	}
}
