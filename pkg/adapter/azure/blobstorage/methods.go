package blobstorage

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"io"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/streaming"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

// GetBlobClient returns the blob storage client
func (bs *blobStorage) GetBlobClient() *azblob.Client {
	return bs.Client
}

// ListBlobs lists all the blobs in a container and returns the contents
func (bs *blobStorage) ListBlobs(ctx context.Context, containerName string) ([]*BlobInfo, error) {
	pager := bs.Client.NewListBlobsFlatPager(containerName, &azblob.ListBlobsFlatOptions{
		Include: azblob.ListBlobsInclude{Snapshots: true, Versions: true},
	})

	var blobs []*BlobInfo
	for pager.More() {
		resp, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing blobs: %s", err.Error())
		}

		if resp.Segment != nil {
			for _, blobFile := range resp.Segment.BlobItems {
				blobs = append(blobs, &BlobInfo{
					Name:         *blobFile.Name,
					FileURL:      fmt.Sprintf("%s/%s", containerName, *blobFile.Name),
					LastModified: *blobFile.Properties.LastModified,
				})
			}
		}
	}

	return blobs, nil
}

// UploadBlobBuffer upload a blob to the service
func (bs *blobStorage) UploadBlobBuffer(ctx context.Context, blobName, containerName string, data []byte) error {
	if _, err := bs.Client.UploadBuffer(ctx, containerName, blobName, data,
		&azblob.UploadBufferOptions{
			AccessConditions: &blob.AccessConditions{
				ModifiedAccessConditions: &blob.ModifiedAccessConditions{IfNoneMatch: to.Ptr(azcore.ETagAny)},
			},
		}); err != nil {
		return fmt.Errorf("error uploading a buffer blob: %s", err.Error())
	}

	return nil
}

// UploadBlobStream upload a blob to the service
func (bs *blobStorage) UploadBlobStream(ctx context.Context, blobName, containerName string, data io.Reader) error {
	if _, err := bs.Client.UploadStream(ctx, containerName, blobName, data,
		&azblob.UploadStreamOptions{
			AccessConditions: &blob.AccessConditions{
				ModifiedAccessConditions: &blob.ModifiedAccessConditions{IfNoneMatch: to.Ptr(azcore.ETagAny)},
			},
		}); err != nil {
		return fmt.Errorf("error uploading a stream blob: %s", err.Error())
	}

	return nil
}

// UploadFile upload a file to the service
func (bs *blobStorage) UploadFile(ctx context.Context, blobName, containerName string, blobSize int) error {
	fileData := make([]byte, blobSize)
	if err := os.WriteFile(blobName, fileData, 0666); err != nil {
		return fmt.Errorf("error writing file: %s", err.Error())
	}

	fileHandler, err := os.Open(blobName)
	if err != nil {
		return fmt.Errorf("error opening the blob file: %s", err.Error())
	}

	defer func(file *os.File) {
		if err = file.Close(); err != nil {
			log.Fatalf("error closing the blob file: %s", err.Error())
		}
	}(fileHandler)

	defer func(name string) {
		if err = os.Remove(name); err != nil {
			log.Fatalf("unnexpected error: %s", err.Error())
		}
	}(blobName)

	if _, err = bs.Client.UploadFile(ctx, containerName, blobName, fileHandler,
		&azblob.UploadFileOptions{
			BlockSize:   int64(1024),
			Concurrency: uint16(3),
		}); err != nil {
		return fmt.Errorf("error uploading the file to blob storage: %s", err.Error())
	}

	return nil
}

// DownloadBlob download a blob from the server and return the content
func (bs *blobStorage) DownloadBlob(ctx context.Context, blobInfo BlobInfo, containerName string) (*azblob.DownloadStreamResponse, error) {
	get, err := bs.Client.DownloadStream(ctx, containerName, blobInfo.Name, &azblob.DownloadStreamOptions{})
	if err != nil {
		return nil, fmt.Errorf("error downloading blob: %s", err.Error())
	}

	return &get, nil
}

// DownloadFile download a file from the server and stores in a file
func (bs *blobStorage) DownloadFile(ctx context.Context, blobInfo BlobInfo, containerName string) error {
	destFile, err := os.Create(blobInfo.Name)
	if err != nil {
		return fmt.Errorf("error creating a file: %s", err.Error())
	}

	defer func(destFile *os.File) {
		if err = destFile.Close(); err != nil {
			log.Fatalf("error closing the blob file: %s", err.Error())
		}
	}(destFile)

	if _, err = bs.Client.DownloadFile(ctx, containerName, blobInfo.Name, destFile, &azblob.DownloadFileOptions{}); err != nil {
		return fmt.Errorf("error downloading blob from blob storage: %s", err.Error())
	}
	return nil
}

// WriteToFile auxiliary function to write to a file from a downloaded stream response
func (bs *blobStorage) WriteToFile(blobName string, response azblob.DownloadStreamResponse) error {
	stream := streaming.NewResponseProgress(
		response.Body,
		func(bytesTransferred int64) {
			fmt.Printf("Downloaded %d bytes.\n", bytesTransferred)
		},
	)

	defer func(stream io.ReadCloser) {
		if err := stream.Close(); err != nil {
			log.Fatalf("error closing the blob file: %s", err.Error())
		}
	}(stream)

	file, err := os.Create(blobName)
	if err != nil {
		return fmt.Errorf("error creating a file: %s", err.Error())
	}

	defer func(file *os.File) {
		if err = file.Close(); err != nil {
			log.Fatalf("error closing the blob file: %s", err.Error())
		}
	}(file)

	if _, err = io.Copy(file, stream); err != nil {
		return fmt.Errorf("error copying a stream to a file: %s", err.Error())
	}

	return nil
}

func (bs *blobStorage) GetSasUrl(blobName, containerName string) (string, error) {
	expiry := time.Now().Add(time.Duration(bs.BlobURLExpiryTime) * time.Minute)
	permissions := sas.BlobPermissions{
		Read: true,
	}

	blobClient := bs.Client.ServiceClient().NewContainerClient(containerName).NewBlobClient(blobName)
	tempURL, err := blobClient.GetSASURL(permissions, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("error creating url to a blob: %s", err.Error())
	}

	return tempURL, nil
}

// CreateBlockBlobClient function to create a block blob client
// it returns a block blob client and an error
// the block blob client is used to upload a file in chunks
// the block blob client is used to mount the file after all the chunks are uploaded
// containerName is a path to the file in the blob storage and start with slash (e.g. /foo/bar)
// fileName is the name of the file in the blob storage (e.g. foo.txt)
// by default the blobURL ends with a slash (e.g. https://foo.blob.core.windows.net/)
func (bs *blobStorage) CreateBlockBlobClient(fileName, containerName string) (*blockblob.Client, error) {
	url := bs.blobUrl
	if url[len(url)-1] == '/' {
		url = url[:len(url)-1]
	}
	blobUrl := fmt.Sprintf("%s/%s/%s", url, containerName, fileName)
	return blockblob.NewClientWithSharedKeyCredential(blobUrl, bs.cred, nil)
}

// PutBlock function to upload a chunk of a file
// it returns the block id and an error
// the block id is used to mount the file after all the chunks are uploaded
func (bs *blobStorage) PutBlock(ctx context.Context, blockBlobClient *blockblob.Client, blockID uint16, data *[]byte) (string, error) {
	base64BlockID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%06d", blockID)))
	_, err := blockBlobClient.StageBlock(ctx, base64BlockID, streaming.NopCloser(bytes.NewReader(*data)), nil)
	if err != nil {
		return "", err
	}

	return base64BlockID, nil
}

// MountFile function to mount a file after all the chunks are uploaded
// it returns an error
// the block ids are used to mount the file
func (bs *blobStorage) MountFile(ctx context.Context, blockBlobClient *blockblob.Client, blockIDs *[]string) error {
	_, err := blockBlobClient.CommitBlockList(ctx, *blockIDs, nil)
	if err != nil {
		return err
	}

	return nil
}

// func (bs *blobStorage) createContainer(ctx context.Context, containerName string) error {
// 	if _, err := bs.Client.CreateContainer(ctx, containerName, nil); err != nil {
// 		return fmt.Errorf("error creating a blob container: %s", err.Error())
// 	}

// 	return nil
// }
