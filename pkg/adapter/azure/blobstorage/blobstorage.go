package blobstorage

import (
	"context"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/faelp22/go-commons-libs/core/config"
)

type BlobInfo struct {
	Name         string
	FileURL      string
	LastModified time.Time
}

type BlobInterface interface {
	GetBlobClient() *azblob.Client
	ListBlobs(ctx context.Context, containerName string) ([]*BlobInfo, error)
	UploadBlobBuffer(ctx context.Context, blobName, containerName string, data []byte) error
	UploadBlobStream(ctx context.Context, blobName, containerName string, data io.Reader) error
	UploadFile(ctx context.Context, blobName, containerName string, blobSize int) error
	DownloadBlob(ctx context.Context, blobInfo BlobInfo, containerName string) (*azblob.DownloadStreamResponse, error)
	DownloadFile(ctx context.Context, blobInfo BlobInfo, containerName string) error
	WriteToFile(blobName string, response azblob.DownloadStreamResponse) error
	GetSasUrl(blobName, containerName string) (string, error)

	// desabilitado
	// createContainer(ctx context.Context, containerName string) error
}

type blobStorage struct {
	Client            *azblob.Client
	BlobURLExpiryTime int64
}

var blobstorage = &blobStorage{}

const DEFAULT_BS_URL_EXPIRY_TIME = 15 // 15 minutes

func New(conf *config.Config) BlobInterface {
	BLOB_STORAGE_ACCOUNT_NAME := os.Getenv("BLOB_STORAGE_ACCOUNT_NAME")
	if BLOB_STORAGE_ACCOUNT_NAME != "" {
		conf.BS_ACCOUNT_NAME = BLOB_STORAGE_ACCOUNT_NAME
	} else {
		log.Println("A variável BLOB_STORAGE_ACCOUNT_NAME é obrigatória!")
		os.Exit(1)
	}

	BLOB_STORAGE_ACCOUNT_KEY := os.Getenv("BLOB_STORAGE_ACCOUNT_KEY")
	if BLOB_STORAGE_ACCOUNT_KEY != "" {
		conf.BS_ACCOUNT_KEY = BLOB_STORAGE_ACCOUNT_KEY
	} else {
		log.Println("A variável BLOB_STORAGE_ACCOUNT_KEY é obrigatória!")
		os.Exit(1)
	}

	BLOB_STORAGE_SERVICE_URL := os.Getenv("BLOB_STORAGE_SERVICE_URL")
	if BLOB_STORAGE_SERVICE_URL != "" {
		conf.BS_SERVICE_URL = BLOB_STORAGE_SERVICE_URL
	} else {
		log.Println("A variável BLOB_STORAGE_SERVICE_URL é obrigatória!")
		os.Exit(1)
	}

	BLOB_STORAGE_EXPIRY_TIME_URL := os.Getenv("BLOB_STORAGE_EXPIRY_TIME_URL")
	if BLOB_STORAGE_EXPIRY_TIME_URL != "" {
		var err error
		conf.BS_URL_EXPIRY_TIME, err = strconv.ParseInt(BLOB_STORAGE_EXPIRY_TIME_URL, 10, 64)
		if err != nil {
			conf.BS_URL_EXPIRY_TIME = DEFAULT_BS_URL_EXPIRY_TIME
		}
	} else {
		conf.BS_URL_EXPIRY_TIME = DEFAULT_BS_URL_EXPIRY_TIME
	}

	if blobstorage == nil || blobstorage.Client == nil {
		cred, err := azblob.NewSharedKeyCredential(conf.BS_ACCOUNT_NAME, conf.BS_ACCOUNT_KEY)
		if err != nil {
			log.Println("Erro criando credencial sharedkey")
			os.Exit(1)
		}

		client, err := azblob.NewClientWithSharedKeyCredential(conf.BS_SERVICE_URL, cred, nil)
		if err != nil {
			log.Println("Erro criando cliente Blob Storage com sharedkey")
			os.Exit(1)
		}

		blobstorage = &blobStorage{
			Client:            client,
			BlobURLExpiryTime: conf.BS_URL_EXPIRY_TIME,
		}
	}

	return blobstorage
}
