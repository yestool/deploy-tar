package buildtar

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
)

func Tar(srcDir, destFile string) {
	err := compress(srcDir, destFile)
	if err != nil {
		log.Fatal(err)
	}
}


func compress(srcDir string, destFile string) error {
	// 创建目标文件
	fw, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer fw.Close()

	// 创建 tar.Writer
	gw := gzip.NewWriter(fw)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Package all the files in the directory
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Ignore the directory itself
		if info.IsDir() && path != srcDir {
			return nil
		}
		// Create a new tar file header
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		// Change the Name field in the file header to a relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		header.Name = relPath
		// Write the file header to tar. Writer
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		// In the case of a file, write the contents of the file to tar. Writer
		if !info.IsDir() {
			fr, err := os.Open(path)
			if err != nil {
				return err
			}
			defer fr.Close()
			_, err = io.Copy(tw, fr)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}