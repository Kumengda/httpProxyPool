package testProxy

import (
	"github.com/tidwall/gjson"
	"golang.org/x/net/html/charset"
	"io"
	"net/http"
)

type Proxy struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func GetProxy(turl string) ([]Proxy, error) {
	var proxyList []Proxy
	client := &http.Client{}
	req, _ := http.NewRequest("GET", turl, nil)
	resp, err := client.Do(req)
	if err != nil {
		return proxyList, err
	}
	defer resp.Body.Close()
	bodyReader, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return proxyList, err
	}

	data, err := io.ReadAll(bodyReader)
	if err != nil {
		return proxyList, err
	}
	jsonData := gjson.Parse(string(data))
	for _, v := range jsonData.Get("data").Array() {
		proxyList = append(proxyList, Proxy{
			Host:     v.Get("host").String(),
			Username: v.Get("username").String(),
			Password: v.Get("password").String(),
		})
	}
	return proxyList, err
}
