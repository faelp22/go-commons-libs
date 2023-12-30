package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/faelp22/go-commons-libs/core/config"
	"github.com/faelp22/go-commons-libs/pkg/adapter/azure/blobstorage"
	"github.com/phuslu/log"
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
		log.Fatal().Msg("filepath is required")
	}

	absFilePath, err := filepath.Abs(*flagFilePath)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	file, err := os.Open(absFilePath)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error().Str("ERRO_BLOB", "error closing file").Msg(err.Error())
		}
	}(file)

	fileInfo, _ := file.Stat()

	log.Debug().Str("FILE_SIZE", unitConverter(int(fileInfo.Size())))
	log.Debug().Str("FILE_NAME", fileInfo.Name())

	fileSize := fileInfo.Size()

	blockBlobClient, err := blobStorageService.CreateBlockBlobClient(fileInfo.Name(), "test")
	if err != nil {
		log.Fatal().Msg(err.Error())
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
			log.Fatal().Msg(err.Error())
		}

		chunkData = chunkData[:n]
		fmt.Printf("upload chunk %d of %d, size %s\r", i, totalNumberOfChunks, unitConverter(len(chunkData)))
		blockID, err := blobStorageService.PutBlock(ctx, blockBlobClient, i, &chunkData)
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		fmt.Println("success to upload chunk, store block ID: ", blockID)
		blockIDs = append(blockIDs, blockID)
	}

	err = blobStorageService.MountFile(ctx, blockBlobClient, &blockIDs)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	log.Info().Msg("Success to upload file")
}
