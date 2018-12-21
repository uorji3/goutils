package goutils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// HTTPReqData ...
type HTTPReqData struct {
	Method, URL, Auth string
	Body, Headers     map[string]string
}

// JSONReqData ...
type JSONReqData struct {
	Method, URL, Auth string
	Body, Headers     map[string]interface{}
}

// HTTPXMLData ...
type HTTPXMLData struct{ Body, URL string }

// Response ...
type Response struct {
	StatusCode   int
	Header       http.Header
	Status, Body string
}

// WritePidFile writes a pid file, but first make sure it doesn't exist with a running pid.
func WritePidFile(pidFile string) error {
	// Read in the pid file as a slice of bytes.
	if piddata, err := ioutil.ReadFile(pidFile); err == nil {
		// Convert the file contents to an integer.
		if pid, err := strconv.Atoi(string(piddata)); err == nil {
			// Look for the pid in the process list.
			if process, err := os.FindProcess(pid); err == nil {
				// Send the process a signal zero kill.
				if err := process.Signal(syscall.Signal(0)); err == nil {
					// We only get an error if the pid isn't running, or it's not ours.
					return fmt.Errorf("pid already running: %d", pid)
				}
			}
		}
	}
	// If we get here, then the pidfile didn't exist,
	// or the pid in it doesn't belong to the user running this app.
	return ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0664)
}

// MakeHTTPRequest ...
func (httpReq *HTTPReqData) MakeHTTPRequest() (Response, error) {
	if len(httpReq.Body) < 0 {
		return Response{}, errors.New("No form body found")
	}
	form := url.Values{}
	for key, value := range httpReq.Body {
		form.Add(key, value)
	}
	client := http.Client{Timeout: time.Second * 60 * 3}
	req, err := http.NewRequest(
		httpReq.Method, httpReq.URL, strings.NewReader(form.Encode()))
	if err != nil {
		return Response{}, fmt.Errorf("makerequest: %v", err)
	}
	req.Header.Add("Content-Length", strconv.Itoa(len(form)))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	if len(httpReq.Headers) > 0 {
		for key, val := range httpReq.Headers {
			req.Header.Add(key, val)
		}
	}
	if len(httpReq.Auth) > 1 {
		user := strings.Split(httpReq.Auth, ":")
		req.SetBasicAuth(user[0], user[1])
	}
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("makerequest do: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("makerequest readall: %v", err)
	}
	return Response{
		Body: string(body), Header: resp.Header, Status: resp.Status, StatusCode: resp.StatusCode,
	}, nil
}

// MakeJSONRequest ...
func (jsonReq *JSONReqData) MakeJSONRequest() (Response, error) {
	jsonStr, err := json.Marshal(jsonReq.Body)
	if err != nil {
		return Response{}, fmt.Errorf("body json marshall: %v", err)
	}
	req, err := http.NewRequest("POST", jsonReq.URL, bytes.NewBuffer(jsonStr))
	if err != nil {
		return Response{}, fmt.Errorf("new request error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if len(jsonReq.Headers) > 0 {
		for key, val := range jsonReq.Headers {
			req.Header.Add(key, val.(string))
		}
	}

	client := http.Client{Timeout: time.Second * 60 * 2}
	if len(jsonReq.Auth) > 1 {
		user := strings.Split(jsonReq.Auth, ":")
		req.SetBasicAuth(user[0], user[1])
	}
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("client do error: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("body read error: %v", err)
	}
	return Response{
		Body: string(body), Header: resp.Header, Status: resp.Status, StatusCode: resp.StatusCode,
	}, nil
}

// MakeXMLRequest ...
func (httpReq *HTTPXMLData) MakeXMLRequest() (Response, error) {
	httpClient := http.Client{Timeout: time.Second * 60 * 2}
	resp, err := httpClient.Post(
		httpReq.URL, "text/xml; charset=utf-8",
		bytes.NewBufferString(httpReq.Body),
	)
	if err != nil {
		return Response{}, fmt.Errorf("xml post error: %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("xml response readall: %v", err)
	}
	return Response{
		Body: string(body), Header: resp.Header, Status: resp.Status, StatusCode: resp.StatusCode,
	}, nil
}
