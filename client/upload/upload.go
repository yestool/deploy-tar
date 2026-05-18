package upload

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/yestool/deploy-tar/client/config"
	"golang.org/x/net/proxy"
)

func UploadTar(config config.Config) error {
	file, err := os.Open(config.TarPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := attachField(writer, "webPath", config.WebPath); err != nil {
		return fmt.Errorf("cannot writeField webPath: %w", err)
	}
	if err := attachField(writer, "apiKey", config.ApiKey); err != nil {
		return fmt.Errorf("cannot writeField apiKey: %w", err)
	}
	if err := attachField(writer, "webSite", config.WebSite); err != nil {
		return fmt.Errorf("cannot writeField website: %w", err)
	}

	// add file
	filePart, err := writer.CreateFormFile("file", filepath.Base(config.TarPath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = io.Copy(filePart, file)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", config.Server, body)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client, err := newHTTPClient(config.Socks5Proxy)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned non-200 status code: %d  body: %s", resp.StatusCode, string(body))
	}

	fmt.Println("File uploaded successfully")
	return nil
}

func newHTTPClient(socks5Proxy string) (*http.Client, error) {
	socks5Proxy = strings.TrimSpace(socks5Proxy)
	if socks5Proxy == "" {
		return &http.Client{}, nil
	}

	proxyAddr, auth, err := parseSocks5Proxy(socks5Proxy)
	if err != nil {
		return nil, err
	}

	dialer, err := proxy.SOCKS5("tcp", proxyAddr, auth, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to create socks5 proxy dialer: %w", err)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = nil
	transport.Dial = dialer.Dial

	return &http.Client{Transport: transport}, nil
}

func parseSocks5Proxy(proxyValue string) (string, *proxy.Auth, error) {
	if !strings.Contains(proxyValue, "://") {
		return proxyValue, nil, nil
	}

	proxyURL, err := url.Parse(proxyValue)
	if err != nil {
		return "", nil, fmt.Errorf("invalid socks5 proxy: %w", err)
	}
	if proxyURL.Scheme != "socks5" && proxyURL.Scheme != "socks5h" {
		return "", nil, fmt.Errorf("unsupported proxy scheme %q, only socks5 is supported", proxyURL.Scheme)
	}
	if proxyURL.Host == "" {
		return "", nil, fmt.Errorf("invalid socks5 proxy: host is required")
	}

	var auth *proxy.Auth
	if proxyURL.User != nil {
		password, _ := proxyURL.User.Password()
		auth = &proxy.Auth{
			User:     proxyURL.User.Username(),
			Password: password,
		}
	}

	return proxyURL.Host, auth, nil
}

func attachField(bodyWriter *multipart.Writer, keyname, keyvalue string) error {
	if err := bodyWriter.WriteField(keyname, keyvalue); err != nil {
		log.Printf("Cannot WriteField: %s, err: %v", keyname, err)
		return err
	}
	return nil
}
