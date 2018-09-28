package goutils

import (
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
	Method, URL   string
	Body, Headers map[string]string
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

// MakeRequest ...
func (httpReq *HTTPReqData) MakeRequest() (string, error) {
	if len(httpReq.Body) < 0 {
		return "", errors.New("No form body found")
	}
	form := url.Values{}
	for key, value := range httpReq.Body {
		form.Add(key, value)
	}
	client := http.Client{Timeout: time.Second * 60 * 3}
	req, err := http.NewRequest(
		httpReq.Method, httpReq.URL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("makerequest: %v", err)
	}
	req.Header.Add("Content-Length", strconv.Itoa(len(form)))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	if len(httpReq.Headers) > 0 {
		for key, val := range httpReq.Headers {
			req.Header.Add(key, val)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("makerequest do: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("makerequest readall: %v", err)
	}
	return string(body), nil
}
