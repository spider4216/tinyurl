package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	endpont := "http://localhost:8080"
	data := url.Values{}

	fmt.Println("Введите длинный URL")
	reader := bufio.NewReader(os.Stdin)
	long, err := reader.ReadString('\n')

	if err != nil {
		fmt.Println("Error", err)
		return
	}

	long = strings.TrimSuffix(long, "\n")
	long = strings.TrimSpace(long)

	data.Set("url", long)

	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
	}

	trans := &http.Transport{
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
	}

	client := &http.Client{
		Transport: trans,
		Timeout:   10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, endpont, strings.NewReader(data.Encode()))

	if err != nil {
		fmt.Println("Error", err)
		return
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Error", err)
		return
	}

	fmt.Println("Статус-код", resp.Status)

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing body: %s", err.Error())
		}
	}()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Error", err)
		return
	}

	fmt.Printf("Response: %s\n", string(body))

}
