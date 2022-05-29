package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
)

func getEnv(key, defaultValue string) string { //Get environment variables from global, or use default
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

var bindAddr = getEnv("LISTEN_HOST", "0.0.0.0:53")
var upstreamAddr = getEnv("UPSTREAM_DNS_HOST", "1.1.1.1:853")
var rootCertsLocation = getEnv("ROOT_CA_LOCATION", "/etc/ssl/cert.pem")
var rootCerts = x509.NewCertPool()

func main() {
	fmt.Printf("Listen : %v\nUpstream dns is : %v\n\n", bindAddr, upstreamAddr)

	listener, err := net.Listen("tcp", bindAddr)
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		log.Println("Handling new connection", conn.RemoteAddr())
		if err != nil {
			log.Fatal(err)
			continue
		}
		go func() {
			defer func(conn net.Conn) {
				err := conn.Close()
				if err != nil {
					log.Fatal(err)
				}
			}(conn)

				if pemBytes, err := ioutil.ReadFile(rootCertsLocation); err == nil {
					rootCerts.AppendCertsFromPEM(pemBytes)
				}

				config := &tls.Config{RootCAs: rootCerts}
				conn2, err := tls.Dial("tcp", upstreamAddr, config)
				if err != nil {
					log.Fatal(err)
				}

			defer func(conn2 net.Conn) {
				err := conn2.Close()
				if err != nil {
					log.Fatal(err)
				}
			}(conn2)

			closer := make(chan struct{}, 2)
			go cp(closer, conn2, conn)
			go cp(closer, conn, conn2)
			<-closer
			log.Println("Connection closed", conn.RemoteAddr())
		}()
	}
}

func cp(closer chan struct{}, dst io.Writer, src io.Reader) {
	_, _ = io.Copy(dst, src)
	closer <- struct{}{}
}
