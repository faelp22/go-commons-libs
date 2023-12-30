package blobstorage

import (
	"context"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/phuslu/log"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"

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
	CreateBlockBlobClient(fileName, containerName string) (*blockblob.Client, error)
	PutBlock(ctx context.Context, blockBlockClient *blockblob.Client, blockID uint16, data *[]byte) (string, error)
	MountFile(ctx context.Context, blockBlobClient *blockblob.Client, blockIDs *[]string) error
	// desabilitado
	// createContainer(ctx context.Context, containerName string) error
}

type blobStorage struct {
	Client            *azblob.Client
	BlobURLExpiryTime int64
	cred              *azblob.SharedKeyCredential
	blobUrl           string
}

var blobstorage = &blobStorage{}

const DEFAULT_BS_URL_EXPIRY_TIME = 15 // 15 minutes

func New(conf *config.Config) BlobInterface {
	BLOB_STORAGE_ACCOUNT_NAME := os.Getenv("BLOB_STORAGE_ACCOUNT_NAME")
	if BLOB_STORAGE_ACCOUNT_NAME != "" {
		conf.BS_ACCOUNT_NAME = BLOB_STORAGE_ACCOUNT_NAME
	} else if conf.AppMode == config.PRODUCTION && conf.AppTargetDeploy == config.TARGET_DEPLOY_NUVEM {
		log.Fatal().Msg("A variável BLOB_STORAGE_ACCOUNT_NAME é obrigatória!")
	}

	BLOB_STORAGE_ACCOUNT_KEY := os.Getenv("BLOB_STORAGE_ACCOUNT_KEY")
	if BLOB_STORAGE_ACCOUNT_KEY != "" {
		conf.BS_ACCOUNT_KEY = BLOB_STORAGE_ACCOUNT_KEY
	} else if conf.AppMode == config.PRODUCTION && conf.AppTargetDeploy == config.TARGET_DEPLOY_NUVEM {
		log.Fatal().Msg("A variável BLOB_STORAGE_ACCOUNT_KEY é obrigatória!")
	}

	BLOB_STORAGE_SERVICE_URL := os.Getenv("BLOB_STORAGE_ACCOUNT_URL")
	if BLOB_STORAGE_SERVICE_URL != "" {
		conf.BS_SERVICE_URL = BLOB_STORAGE_SERVICE_URL
	} else if conf.AppMode == config.PRODUCTION && conf.AppTargetDeploy == config.TARGET_DEPLOY_NUVEM {
		log.Fatal().Msg("A variável BLOB_STORAGE_ACCOUNT_URL é obrigatória!")
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
			log.Fatal().Msg("Erro criando credencial sharedkey!")
		}

		client, err := azblob.NewClientWithSharedKeyCredential(conf.BS_SERVICE_URL, cred, nil)
		if err != nil {
			log.Fatal().Msg("Erro criando cliente Blob Storage com sharedkey!")
		}

		blobstorage = &blobStorage{
			Client:            client,
			BlobURLExpiryTime: conf.BS_URL_EXPIRY_TIME,
			cred:              cred,
			blobUrl:           BLOB_STORAGE_SERVICE_URL,
		}
	}

	return blobstorage
}
