package urlutil

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Check URL is reachable
func IsURLReachable(url string) error {
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Head(url)
	if err != nil {
		return errors.New("Unable to access the URL")
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return errors.New("Server responded with non 200 status code")
	}
}

// Check Server is reachable
// Does not validate the status code
func IsServerReachable(serverURL string) error {
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Head(serverURL)
	if err != nil {

		if urlErr, ok := err.(*url.Error); ok {
			if urlErr.Timeout() {
				return errors.New("Server connection timeout")
			}

			errMsg := urlErr.Err.Error()
			if strings.Contains(errMsg, "no such host") {
				return errors.New("Hostname not found")
			} else if strings.Contains(errMsg, "connection refused") {
				return errors.New("connection refused or host unreachable")
			} else if strings.Contains(errMsg, "network is unreachable") {
				return errors.New("network is unreachable")
			}
		}

		if _, ok := err.(*net.DNSError); ok {
			return errors.New("Hostname not found")
		}

		if netErr, ok := err.(net.Error); ok {
			if netErr.Timeout() {
				return errors.New("Server connection timeout")
			} else if _, ok := netErr.(*net.OpError); ok {
				return errors.New("connection refused or host unreachable")
			}
		}

		return err
	}

	defer resp.Body.Close()

	return nil
}
