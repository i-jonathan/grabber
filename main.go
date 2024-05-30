package main

import (
	"fmt"
	"log"
	"os"
	"sync"
)

var fileIsh FileData

func main() {
	//start := time.Now()
	var urlArg = "http://localhost:8560/The%20Hunger%20Games%20The%20Ballad%20Of%20Songbirds%20Snakes%20%282023%29%20%5B1080p%5D%20%5BWEBRip%5D%20%5B5.1%5D%20%5BYTS.MX%5D.zip"
	//fmt.Print("URL: ")
	//_, err := fmt.Scanln(&urlArg)
	//if err != nil {
	//	log.Print(err)
	//	os.Exit(1)
	//}
	workers := 1
	//fmt.Println("Checking file at URl: ", urlArg)
	evaluateHeader(urlArg)

	//var writeWg sync.WaitGroup
	//writeQueue := make(chan WriterInfo, 1000)
	//writeWg.Add(1)
	//go writeToFile(writeQueue, &writeWg)

	if fileIsh.Size < 1 {
		var writeWg sync.WaitGroup
		writeQueue := make(chan WriterInfo, 1000)
		writeWg.Add(1)
		go writeToFile(writeQueue, &writeWg)
		fmt.Println("Couldn't get file size. Unable to concurrently download file.")
		err := regularDownloader(urlArg, writeQueue)
		if err != nil {
			log.Println(err)
			return
		}
		writeWg.Wait()
		close(writeQueue)
		fmt.Println("Download Complete")
		os.Exit(0)
	}

	concurrentDownload(urlArg, fileIsh.Size, workers)
	//close(writeQueue)
	//writeWg.Wait()
	//fmt.Println("Download Complete")
	//endTime := time.Now()
	//elapsed := endTime.Sub(start)
	//fmt.Println("Elapsed time: ", elapsed)
	//fmt.Printf("Average Speed: %.2fmb/s\n", float64(fileIsh.Size/1048576)/elapsed.Seconds())
}

func concurrentDownload(link string, size int64, workers int) {
	//fmt.Println("Creating file...")
	_ = os.Remove(fileIsh.Name)
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
			err := downloadChunk(link, start, chunkSize, &downWg)
			if err != nil {
				fmt.Println("Error downloading chunk: ", err)
			}
		}()
	}
	downWg.Wait()
}
