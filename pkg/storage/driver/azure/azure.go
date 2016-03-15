package azure

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	storagedriver "github.com/deis/builder/pkg/storage/driver"
	"github.com/deis/builder/pkg/storage/driver/base"
	"github.com/deis/builder/pkg/storage/driver/factory"

	azure "github.com/Azure/azure-sdk-for-go/storage"
)

const driverName = "azure"

const (
	paramAccountName = "accountname"
	paramAccountKey  = "accountkey"
	paramContainer   = "builder-container"
	paramRealm       = "realm"
	maxChunkSize     = 4 * 1024 * 1024
)

type driver struct {
	client    azure.BlobStorageClient
	container string
}

type baseEmbed struct{ base.Base }

// Driver is a storagedriver.StorageDriver implementation backed by
// Microsoft Azure Blob Storage Service.
type Driver struct{ baseEmbed }

func init() {
	factory.Register(driverName, &azureDriverFactory{})
}

type azureDriverFactory struct{}

func (factory *azureDriverFactory) Create(parameters map[string]string) (storagedriver.StorageDriver, error) {
	return FromParameters(parameters)
}

// FromParameters constructs a new Driver with a given parameters map.
func FromParameters(parameters map[string]string) (*Driver, error) {
	accountName, ok := parameters[paramAccountName]
	if !ok || fmt.Sprint(accountName) == "" {
		return nil, fmt.Errorf("No %s parameter provided", paramAccountName)
	}

	accountKey, ok := parameters[paramAccountKey]
	if !ok || fmt.Sprint(accountKey) == "" {
		return nil, fmt.Errorf("No %s parameter provided", paramAccountKey)
	}

	container, ok := parameters[paramContainer]
	if !ok || fmt.Sprint(container) == "" {
		return nil, fmt.Errorf("No %s parameter provided", paramContainer)
	}

	realm, ok := parameters[paramRealm]
	if !ok || fmt.Sprint(realm) == "" {
		realm = azure.DefaultBaseURL
	}

	return New(fmt.Sprint(accountName), fmt.Sprint(accountKey), fmt.Sprint(container), fmt.Sprint(realm))
}

// New constructs a new Driver with the given Azure Storage Account credentials
func New(accountName, accountKey, container, realm string) (*Driver, error) {
	api, err := azure.NewClient(accountName, accountKey, realm, azure.DefaultAPIVersion, true)
	if err != nil {
		return nil, err
	}

	blobClient := api.GetBlobService()

	// Create registry container
	if _, err = blobClient.CreateContainerIfNotExists(container, azure.ContainerAccessTypePrivate); err != nil {
		return nil, err
	}

	d := &driver{
		client:    blobClient,
		container: container}
	return &Driver{baseEmbed: baseEmbed{Base: base.Base{StorageDriver: d}}}, nil
}

// Implement the storagedriver.StorageDriver interface.
func (d *driver) Name() string {
	return driverName
}

func (d *driver) CheckConnectionStatus() (bool, error) {
	_, err := d.client.ListContainers(azure.ListContainersParameters{})
	if err != nil {
		return false, err
	}
	return true, err
}

// GetContent retrieves the content stored at "path" as a []byte.
func (d *driver) GetContent(path string) ([]byte, error) {
	blob, err := d.client.GetBlob(d.container, path)
	if err != nil {
		if is404(err) {
			return nil, storagedriver.PathNotFoundError{Path: path}
		}
		return nil, err
	}

	return ioutil.ReadAll(blob)
}

// PutContent stores the []byte content at a location designated by "path".
func (d *driver) PutContent(path string, contents []byte) error {
	if _, err := d.client.DeleteBlobIfExists(d.container, path); err != nil {
		return err
	}
	writer, err := d.Writer(path, false)
	if err != nil {
		return err
	}
	defer writer.Close()
	_, err = writer.Write(contents)
	if err != nil {
		return err
	}
	return writer.Commit()
}

// Writer returns a FileWriter which will store the content written to it
// at the location designated by "path" after the call to Commit.
func (d *driver) Writer(path string, append bool) (storagedriver.FileWriter, error) {
	blobExists, err := d.client.BlobExists(d.container, path)
	if err != nil {
		return nil, err
	}
	var size int64
	if blobExists {
		if append {
			blobProperties, err := d.client.GetBlobProperties(d.container, path)
			if err != nil {
				return nil, err
			}
			size = blobProperties.ContentLength
		} else {
			err := d.client.DeleteBlob(d.container, path)
			if err != nil {
				return nil, err
			}
		}
	} else {
		if append {
			return nil, storagedriver.PathNotFoundError{Path: path}
		}
		err := d.client.PutAppendBlob(d.container, path, nil)
		if err != nil {
			return nil, err
		}
	}

	return d.newWriter(path, size), nil
}

// Stat retrieves the FileInfo for the given path, including the current size
// in bytes and the creation time.
func (d *driver) Stat(path string) (storagedriver.FileInfo, error) {
	// Check if the path is a blob
	if ok, err := d.client.BlobExists(d.container, path); err != nil {
		return nil, err
	} else if ok {
		blob, err := d.client.GetBlobProperties(d.container, path)
		if err != nil {
			return nil, err
		}

		mtim, err := time.Parse(http.TimeFormat, blob.LastModified)
		if err != nil {
			return nil, err
		}

		return storagedriver.FileInfoInternal{FileInfoFields: storagedriver.FileInfoFields{
			Path:    path,
			Size:    int64(blob.ContentLength),
			ModTime: mtim,
			IsDir:   false,
		}}, nil
	}

	// Check if path is a virtual container
	virtContainerPath := path
	if !strings.HasSuffix(virtContainerPath, "/") {
		virtContainerPath += "/"
	}
	blobs, err := d.client.ListBlobs(d.container, azure.ListBlobsParameters{
		Prefix:     virtContainerPath,
		MaxResults: 1,
	})
	if err != nil {
		return nil, err
	}
	if len(blobs.Blobs) > 0 {
		// path is a virtual container
		return storagedriver.FileInfoInternal{FileInfoFields: storagedriver.FileInfoFields{
			Path:  path,
			IsDir: true,
		}}, nil
	}

	// path is not a blob or virtual container
	return nil, storagedriver.PathNotFoundError{Path: path}
}

func is404(err error) bool {
	statusCodeErr, ok := err.(azure.AzureStorageServiceError)
	return ok && statusCodeErr.StatusCode == http.StatusNotFound
}

type writer struct {
	driver    *driver
	path      string
	size      int64
	bw        *bufio.Writer
	closed    bool
	committed bool
	cancelled bool
}

func (d *driver) newWriter(path string, size int64) storagedriver.FileWriter {
	return &writer{
		driver: d,
		path:   path,
		size:   size,
		bw: bufio.NewWriterSize(&blockWriter{
			client:    d.client,
			container: d.container,
			path:      path,
		}, maxChunkSize),
	}
}

func (w *writer) Write(p []byte) (int, error) {
	if w.closed {
		return 0, fmt.Errorf("already closed")
	} else if w.committed {
		return 0, fmt.Errorf("already committed")
	} else if w.cancelled {
		return 0, fmt.Errorf("already cancelled")
	}

	n, err := w.bw.Write(p)
	w.size += int64(n)
	return n, err
}

func (w *writer) Size() int64 {
	return w.size
}

func (w *writer) Close() error {
	if w.closed {
		return fmt.Errorf("already closed")
	}
	w.closed = true
	return w.bw.Flush()
}

func (w *writer) Cancel() error {
	if w.closed {
		return fmt.Errorf("already closed")
	} else if w.committed {
		return fmt.Errorf("already committed")
	}
	w.cancelled = true
	return w.driver.client.DeleteBlob(w.driver.container, w.path)
}

func (w *writer) Commit() error {
	if w.closed {
		return fmt.Errorf("already closed")
	} else if w.committed {
		return fmt.Errorf("already committed")
	} else if w.cancelled {
		return fmt.Errorf("already cancelled")
	}
	w.committed = true
	return w.bw.Flush()
}

type blockWriter struct {
	client    azure.BlobStorageClient
	container string
	path      string
}

func (bw *blockWriter) Write(p []byte) (int, error) {
	n := 0
	for offset := 0; offset < len(p); offset += maxChunkSize {
		chunkSize := maxChunkSize
		if offset+chunkSize > len(p) {
			chunkSize = len(p) - offset
		}
		err := bw.client.AppendBlock(bw.container, bw.path, p[offset:offset+chunkSize])
		if err != nil {
			return n, err
		}

		n += chunkSize
	}

	return n, nil
}
