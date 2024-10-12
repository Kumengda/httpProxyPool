package httpProxyPool

import (
	"crypto/tls"
	"fmt"
	"golang.org/x/net/proxy"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type HttpProxyClientInfo struct {
	hashCode   uint32
	HttpClient *http.Client
	UsingNum   int32
	ProxyUrl   string
}

func NewHttpProxyClientInfo(username, password, host string, timeout int) HttpProxyClientInfo {
	return HttpProxyClientInfo{
		HttpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
			Transport: &http.Transport{
				Proxy: func(req *http.Request) (*url.URL, error) {
					return url.Parse(fmt.Sprintf("http://%s:%s@%s", username, password, host))
				},
			},
		},
		ProxyUrl: host,
		hashCode: generateRandomHashCode(),
	}
}
func NewHttpSockets5ClientInfo(username, password, host string, timeout int) (HttpProxyClientInfo, error) {
	auth := proxy.Auth{
		User:     username,
		Password: password,
	}
	dialer, err := proxy.SOCKS5("tcp", host, &auth, proxy.Direct)
	if err != nil {
		fmt.Println("Error creating dialer:", err)
		return HttpProxyClientInfo{}, err
	}

	return HttpProxyClientInfo{
		HttpClient: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
			Transport: &http.Transport{
				Dial: dialer.Dial,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
		ProxyUrl: host,
		hashCode: generateRandomHashCode(),
	}, nil
}

func generateRandomHashCode() uint32 {
	rander := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rander.Uint32()
}
