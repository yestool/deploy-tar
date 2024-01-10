package upload

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"github.com/yestool/deploy-tar/client/config"
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

	client := http.Client{}
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


func attachField(bodyWriter * multipart.Writer, keyname, keyvalue string) error {
	if err := bodyWriter.WriteField(keyname, keyvalue); err != nil {
			log.Printf("Cannot WriteField: %s, err: %v", keyname, err)
			return err
	}
	return nil
}