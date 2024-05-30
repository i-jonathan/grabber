package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
	"time"
)

type FileData struct {
	Size int64
	Name string
}

type WriterInfo struct {
	Offset    int64
	Data      []byte
	ByteCount int64
}

func evaluateHeader(targetUrl string) {
	httpClient := &http.Client{Timeout: 60 * time.Second}
	resp, err := httpClient.Head(targetUrl)
	if err != nil {
		log.Println(err)
		fileIsh.Size = -1
	}

	fileIsh.Size = resp.ContentLength
	fileIsh.Name = getFileName(targetUrl)

	return
}

func getFileName(targetUrl string) string {
	// consider using content disposition maybe with regex
	tUrl, err := url.Parse(targetUrl)
	if err != nil {
		log.Println(err)
		panic("Ah!")
	}

	return path.Base(tUrl.Path)
}

func createFile() bool {
	file, err := os.Create(fileIsh.Name)
	if err != nil {
		log.Println(err)
		return false
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)

	err = file.Truncate(fileIsh.Size)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func writeToFile(writeQueue <-chan WriterInfo, wg *sync.WaitGroup) {
	// concurrent function to write to file
	defer wg.Done()
	file, err := os.OpenFile(fileIsh.Name, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file: ", err)
		return
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)

	for value := range writeQueue {
		_, err := file.Seek(value.Offset, io.SeekStart)
		if err != nil {
			log.Println("Error when writing to file", err)
			return
		}
		writer := bufio.NewWriter(file)
		_, err = writer.Write(value.Data)
		if err != nil {
			fmt.Println("Error writing file: ", err)
			return
		}
		if err := writer.Flush(); err != nil {
			log.Println("Error writing file: ", err)
			return
		}

	}
}

func downloadChunk(targetUrl string, offset int64, chunkSize int64, wg *sync.WaitGroup) error {
	// concurrently download parts of file
	defer wg.Done()

	req, err := http.NewRequest(http.MethodGet, targetUrl, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, offset+chunkSize-1))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			log.Println(err)
		}
	}(resp.Body)

	// testing writing to file directly
	file, err := os.OpenFile(fileIsh.Name, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file: ", err)
		return err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)

	_, err = file.Seek(offset, io.SeekStart)
	if err != nil {
		fmt.Println("Error when seeking to offset", err)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println("Error when writing to file", err)
	}
	return nil
}

func regularDownloader(targetUrl string, writeQueue chan<- WriterInfo) error {
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, targetUrl, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(resp.Body)
	var offset int64 = 0
	for {
		buffer := make([]byte, 1024*1024*1024)
		n, err := reader.Read(buffer)
		if err != nil {
			return err
		}

		if n == 0 {
			break
		}

		writeQueue <- WriterInfo{Offset: offset, Data: buffer}
		offset += int64(n)
	}

	return nil
}
