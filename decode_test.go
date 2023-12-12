package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDecode(t *testing.T) {

	go Serve()

	url := "http://localhost:3000/api/user/SID"
	method := "POST"

	payload := strings.NewReader("name=John%20Doe&email=john.doe%40example.com&age=19")

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
