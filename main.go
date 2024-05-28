package main

import (
	"fmt"
	"log"
	"os"
	"sync"
)

var fileIsh FileData

func main() {
	var urlArg string
	fmt.Print("URL: ")
	_, err := fmt.Scanln(&urlArg)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	workers := 5
	fmt.Println("Checking file at URl: ", urlArg)
	evaluateHeader(urlArg)

	var writeWg sync.WaitGroup
	writeQueue := make(chan WriterInfo, 1000)
	writeWg.Add(1)
	go writeToFile(writeQueue, &writeWg)

	if fileIsh.Size < 1 {
		fmt.Println("Couldn't get file size. Unable to concurrently download file.")
		err := regularDownloader(urlArg, writeQueue)
		if err != nil {
			log.Println(err)
			return
		}
		writeWg.Wait()
		fmt.Println("Download Complete")
		os.Exit(0)
	}

	concurrentDownload(urlArg, fileIsh.Size, workers, writeQueue)
	close(writeQueue)
	writeWg.Wait()
	fmt.Println("Download Complete")
}

func concurrentDownload(link string, size int64, workers int, writeQueue chan<- WriterInfo) {
	fmt.Println("Creating file...")
	if ok := createFile(); !ok {
		os.Exit(1)
	}

	var downWg sync.WaitGroup
	chunkSize := size / int64(workers)
	for i := 0; i < workers; i++ {
		start := chunkSize * int64(i)
		end := start + chunkSize - 1
		if end > fileIsh.Size {
			end = fileIsh.Size - 1
		}

		downWg.Add(1)
		go func() {
			err := downloadChunk(link, start, chunkSize, writeQueue, &downWg)
			if err != nil {
				fmt.Println("Error downloading chunk: ", err)
			}
		}()
	}
	downWg.Wait()
}
