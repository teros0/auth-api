package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestGet(t *testing.T) {
	var tests = []struct {
		key  string
		code int
	}{
		{"123", 200},
		{"222", 200},
		{"asdf", 401},
		{"12345", 401},
	}
	client := &http.Client{}
	url := "http://localhost:8000/get-token/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	for _, test := range tests {
		req.Header.Set("Key", test.key)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != test.code {
			t.Errorf("GetToken(%s) returned %d", test.key, test.code)
		}
	}
}

func TestVer(t *testing.T) {
	var tests = []struct {
		token string
		code  int
	}{
		{"321", 200},
		{"333", 200},
		{"asdf", 400},
		{"12345", 400},
	}
	client := &http.Client{}
	url := "http://localhost:8000/ver-token/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	for _, test := range tests {
		req.Header.Set("Token", test.token)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != test.code {
			t.Errorf("VerToken(%s) returned %d", test.token, test.code)
		}
	}
}

func TestDel(t *testing.T) {
	var tests = []struct {
		token string
		code  int
	}{
		{"321", 200},
		{"333", 200},
		{"asdf", 400},
		{"12345", 400},
		{"321", 400}, // Deleting previously deleted token
		{"333", 400},
	}
	client := &http.Client{}
	url := "http://localhost:8000/del-token/"
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	for _, test := range tests {
		req.Header.Set("Token", test.token)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != test.code {
			t.Errorf("DelToken(%s) returned %d, not %d", test.token, resp.StatusCode, test.code)
		}
	}
}