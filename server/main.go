package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"github.com/spf13/viper"
)
const MaxUploadSize = 1024 * 1024 * 100 // 100 MB = 104857600 B
var (
	API_KEY string
	SERVER_PORT int
	TARROOT string
	KEEPFILES int
	MAX_UPLOAD_SIZE int64
	UPLOAD_HANDLE string
)





func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	err := viper.ReadInConfig()
	if err != nil {
		log.Println("Can not read config:", viper.ConfigFileUsed())
	}
	
	API_KEY = viper.GetString("apiKey")
	SERVER_PORT = viper.GetInt("serverPort")
	if (SERVER_PORT == 0) {
		SERVER_PORT = 8080
	}
	KEEPFILES = viper.GetInt("keepFiles")
	MAX_UPLOAD_SIZE = viper.GetInt64("maxUploadSize")
	if (MAX_UPLOAD_SIZE == 0) {
		MAX_UPLOAD_SIZE = MaxUploadSize
	}
	TARROOT = viper.GetString("tarPath")
	if (TARROOT == "") {
		TARROOT = "./uploadfiles"
	}
	UPLOAD_HANDLE = viper.GetString("funcHandle")
	if (UPLOAD_HANDLE == "") {
		UPLOAD_HANDLE = "/upload"
	}
	mux := http.NewServeMux()
	mux.HandleFunc(UPLOAD_HANDLE, uploadHandler)
	log.Printf("Server listen on port %d", SERVER_PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", SERVER_PORT), mux))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
			http.Error(w, fmt.Sprintf("File size need to be less than %d B", MAX_UPLOAD_SIZE), http.StatusBadRequest)
	}
	webPath := r.FormValue("webPath")
	if webPath == "" {
		http.Error(w, "webPath is required", http.StatusBadRequest)
		return
	}
	webSite := r.FormValue("webSite")
	if webSite == "" {
		http.Error(w, "webSite is required", http.StatusBadRequest)
		return
	}
	apiKey := r.FormValue("apiKey")
	if apiKey == "" {
		http.Error(w, "apiKey is required", http.StatusBadRequest)
		return
	}
	if apiKey != API_KEY {
		http.Error(w, "apiKey is invalid", http.StatusBadRequest)
		return
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from request", http.StatusBadRequest)
		return
	}
	fmt.Printf("the uploaded file: name[%s], size[%d], website[%s]\n", fileHeader.Filename, fileHeader.Size, webSite)
	defer file.Close()
	now := time.Now().Format("20060102150405")
	newFileName := now + "-" + fileHeader.Filename
	err = os.MkdirAll(fmt.Sprintf("%s/%s", TARROOT, webSite), os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	newTarFilePath := fmt.Sprintf("%s/%s/%s", TARROOT, webSite, newFileName)
	f, err := os.Create(newTarFilePath)
	if err != nil {
		http.Error(w, "Failed to create file on server", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		http.Error(w, "Failed to save file on server", http.StatusInternalServerError)
		return
	}
	err = unTarGz(newTarFilePath, webPath)
	if err != nil {
		http.Error(w, "Failed to extract file", http.StatusInternalServerError)
		return
	}
	archive(webSite)
	fmt.Fprint(w, "File uploaded and extracted successfully")
}

func unTarGz(tarGzFilePath string, destDir string) error {
	f, err := os.Open(tarGzFilePath)
	if err != nil {
		return fmt.Errorf("failed to open archive file: %v", err)
	}
	defer f.Close()
	gzf, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzf.Close()
	tr := tar.NewReader(gzf)
	// Extract files from the archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %v", err)
		}
		targetPath := filepath.Join(destDir, header.Name)
		if header.FileInfo().Mode().IsDir() {
			// Create directory if it doesn't exist
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
			continue
		}
		// Create parent directory if it doesn't exist
		parentDir := filepath.Dir(targetPath)
		if _, err := os.Stat(parentDir); os.IsNotExist(err) {
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				return fmt.Errorf("failed to create parent directory: %v", err)
			}
		}
		// Extract file
		file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, header.FileInfo().Mode())
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		if _, err := io.Copy(file, tr); err != nil {
			file.Close()
			return fmt.Errorf("failed to extract file: %v", err)
		}
		file.Close()
	}
	return nil
}

func archive(webSite string)  {
	dir := fmt.Sprintf("%s/%s", TARROOT, webSite)
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	type FileInfo struct {
		Name string
		Time time.Time
	}
	
	fileInfos := make([]FileInfo, 0)
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "20") && strings.HasSuffix(file.Name(), ".tar.gz") {
			timestamp := file.Name()[:14]
			t, err := time.Parse("20060102150405", timestamp)
			if err != nil {
				log.Println(err)
				continue
			}
			fileInfos = append(fileInfos, FileInfo{Name: file.Name(), Time: t})
		}
	}
	
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].Time.After(fileInfos[j].Time)
	})
	if KEEPFILES != 0 && len(fileInfos) > KEEPFILES {
		for i := KEEPFILES; i < len(fileInfos); i++ {
			filePath := filepath.Join(dir, fileInfos[i].Name)
			err := os.Remove(filePath)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

