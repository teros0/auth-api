package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

type JSONResp struct {
	Token string `json:"token"`
	Value `json:"value"`
}

type Value struct {
	AccessKey string `json:"access_key"`
}

func TestGet(t *testing.T) {
	var tests = []struct {
		key  string
		code int
	}{
		{"1key", 200},
		{"2key", 200},
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

func TestGetJSON(t *testing.T) {
	var tests = []struct {
		key   string
		token string
	}{
		{"1key", "1token"},
		{"2key", "2token"},
		{"3key", "3token"},
		{"4key", "4token"},
		{"asfdg", ""},
		{"4kee", ""},
		{"wrong", ""},
		{"lkbdsf", ""},
	}

	client := &http.Client{}
	url := "http://localhost:8000/get-token/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}

	for _, test := range tests {
		jsonResp := JSONResp{}
		req.Header.Set("Key", test.key)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		r, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(r, &jsonResp)
		if err != nil {
			t.Error(err)
		}
		if jsonResp.AccessKey != test.key {
			if jsonResp.AccessKey != "" {
				t.Errorf("Sent access key (%s) got %s", test.key, jsonResp.AccessKey)
			}
		}

		if jsonResp.Token != test.token {
			t.Errorf("Sent token (%s) got %s", test.token, jsonResp.Token)
		}
	}
}

func TestVer(t *testing.T) {
	var tests = []struct {
		token string
		code  int
	}{
		{"1token", 200},
		{"2token", 200},
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

func TestVerJSON(t *testing.T) {
	var tests = map[string]string{
		"1token":    "1key",
		"2token":    "2key",
		"3token":    "3key",
		"4token":    "4key",
		"dsfbd":     "",
		"4to124ken": "",
		"4toweken":  "",
	}
	client := &http.Client{}
	url := "http://localhost:8000/ver-token/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	for token, key := range tests {
		val := Value{}
		req.Header.Set("Token", token)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		r, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(r, &val)

		if val.AccessKey != key {
			t.Errorf("Sent token %s, expected key %s got %s", token, key, val.AccessKey)
		}
	}
}

func TestDel(t *testing.T) {
	var tests = []struct {
		token string
		code  int
	}{
		{"1token", 200},
		{"2token", 200},
		{"asdf", 400},
		{"12345", 400},
		{"2key", 400}, // Deleting previously deleted token
		{"1key", 400},
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

var client = &http.Client{}
var url = "http://localhost:8000/get-token/"
var req = func() *http.Request {
	r, _ := http.NewRequest("GET", url, nil)
	r.Header.Set("Key", "123")
	return r
}()

func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		resp, _ := client.Do(req)
		defer resp.Body.Close()
	}
}

var url1 = "http://localhost:8000/ver-token/"
var req1 = func() *http.Request {
	r, _ := http.NewRequest("GET", url1, nil)
	r.Header.Set("Token", "321")
	return r
}()

func BenchmarkVer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		resp, _ := client.Do(req1)
		defer resp.Body.Close()
	}
}
