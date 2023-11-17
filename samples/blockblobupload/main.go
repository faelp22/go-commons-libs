package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/faelp22/go-commons-libs/pkg/adapter/azure/blobstorage"
	"log"
	"os"
	"path/filepath"
)

const (
	_      = iota
	KB int = 1 << (10 * iota)
	MB
	GB
	TB
	PB
)

func unitConverter(size int) string {
	unit := ""
	value := float64(size)
	switch {
	case size >= PB:
		unit = "PB"
		value = float64(size / PB)
	case size >= TB:
		unit = "TB"
		value = float64(size / TB)
	case size >= GB:
		unit = "GB"
		value = float64(size / GB)
	case size >= MB:
		unit = "MB"
		value = float64(size / MB)
	case size >= KB:
		unit = "KB"
		value = float64(size / KB)
	case size >= 0:
		unit = "B"
		value = float64(size)
	default:
		unit = "?"
	}

	return fmt.Sprintf("%.2f %s", value, unit)
}

func main() {
	conf := &config.Config{
		BlobStorage: &config.BlobStorage{},
	}
	blobStorageService := blobstorage.New(conf)

	flagFilePath := flag.String("filepath", "", "file path")
	flag.Parse()

	if *flagFilePath == "" {
		log.Println("filepath is required")
		os.Exit(1)
	}

	absFilePath, err := filepath.Abs(*flagFilePath)
	if err != nil {
		panic(err)
	}

	file, err := os.Open(absFilePath)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println("error closing file: ", err.Error())
		}
	}(file)

	fileInfo, _ := file.Stat()

	log.Println("file size: ", unitConverter(int(fileInfo.Size())))
	log.Println("file name: ", fileInfo.Name())

	fileSize := fileInfo.Size()

	blockBlobClient, err := blobStorageService.CreateBlockBlobClient(fileInfo.Name(), "test")
	if err != nil {
		panic(err)
	}

	chunkSize := 4 << 20                                                                // 4MB
	totalNumberOfChunks := uint16((fileSize + int64(chunkSize) - 1) / int64(chunkSize)) // round up
	if totalNumberOfChunks == 0 {
		totalNumberOfChunks++
	}

	ctx := context.Background()

	blockIDs := make([]string, 0, totalNumberOfChunks)

	for i := uint16(1); i <= totalNumberOfChunks; i++ {
		chunkData := make([]byte, chunkSize)
		n, err := file.Read(chunkData)
		if err != nil {
			panic(err)
		}

		chunkData = chunkData[:n]
		fmt.Printf("upload chunk %d of %d, size %s\r", i, totalNumberOfChunks, unitConverter(len(chunkData)))
		blockID, err := blobStorageService.PutBlock(ctx, blockBlobClient, i, &chunkData)
		if err != nil {
			panic(err)
		}
		fmt.Println("success to upload chunk, store block ID: ", blockID)
		blockIDs = append(blockIDs, blockID)
	}

	err = blobStorageService.MountFile(ctx, blockBlobClient, &blockIDs)
	if err != nil {
		panic(err)
	}

	log.Println("Success to upload file")
}
