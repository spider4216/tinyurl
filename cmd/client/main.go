package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	endpont := "http://localhost:8080"
	data := url.Values{}

	fmt.Println("Введите длинный URL")
	reader := bufio.NewReader(os.Stdin)
	long, err := reader.ReadString('\n')

	if err != nil {
		panic(err)
	}

	long = strings.TrimSuffix(long, "\n")
	long = strings.TrimSpace(long)

	data.Set("url", long)

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPost, endpont, strings.NewReader(data.Encode()))

	if err != nil {
		panic(err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	fmt.Println("Статус-код", resp.Status)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	fmt.Printf("Response: %s\n", string(body))

}
