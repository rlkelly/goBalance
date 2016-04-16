package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

type Registry map[string][]string

var ServiceRegistry = Registry{
	"serviceone/v1": {
		"localhost:9091",
		"localhost:9092",
	},
}

func extractNameVersion(target *url.URL) (name, version string, err error) {
	path := target.Path
	// Trim the leading `/`
	if len(path) > 1 && path[0] == '/' {
		path = path[1:]
	}
	// Explode on `/` and make sure we have at least
	// 2 elements (service name and version)
	tmp := strings.Split(path, "/")
	if len(tmp) < 2 {
		return "", "", fmt.Errorf("Invalid path")
	}
	name, version = tmp[0], tmp[1]
	// Rewrite the request's path without the prefix.
	target.Path = "/" + strings.Join(tmp[2:], "/")
	return name, version, nil
}

// NewMultipleHostReverseProxy creates a reverse proxy that will randomly
// select a host from the passed `targets`
func NewMultipleHostReverseProxy(reg Registry) *httputil.ReverseProxy {
	go func() {
		for {
			time.Sleep(10 * time.Second)
			reg["serviceone/v1"] = CheckPorts(9091, 9099)
			fmt.Println(reg)
		}

		fmt.Println()
	}()
	director := func(req *http.Request) {
		name, version, err := extractNameVersion(req.URL)
		if err != nil {
			log.Print(err)
			return
		}
		endpoints := reg[name+"/"+version]
		if len(endpoints) == 0 {
			log.Printf("Service/Version not found")
			return
		}
		req.URL.Scheme = "http"
		req.URL.Host = name + "/" + version
	}
	return &httputil.ReverseProxy{
		Director: director,
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return http.ProxyFromEnvironment(req)
			},
			Dial: func(network, addr string) (net.Conn, error) {
				addr = strings.Split(addr, ":")[0]
				endpoints := reg[addr]
				if len(endpoints) == 0 {
					return nil, fmt.Errorf("Service/Version not found")
				}
				for {
					fmt.Println(network, endpoints)
					url, err := net.Dial(network, endpoints[rand.Int()%len(endpoints)])
					if err == nil {
						return url, err
					}
				}
			},
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

func CheckPorts(start int, stop int) []string {

	var validports []string

	for val := start; val < stop; val++ {

		var port bytes.Buffer

		port.WriteString("localhost:")
		port.WriteString(strconv.Itoa(val))
		_, err := net.Dial("tcp", port.String())
		if err == nil {
			validports = append(validports, port.String())
		}
	}
	return validports
}

func main() {
	proxy := NewMultipleHostReverseProxy(Registry{
		"serviceone/v1": CheckPorts(9091, 9099),
		"serviceone/v2": {"localhost:9095"},
	})
	log.Fatal(http.ListenAndServe(":9090", proxy))
}
