package gcs

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/googleapi"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"

	storagedriver "github.com/deis/builder/pkg/storage/driver"
	"github.com/deis/builder/pkg/storage/driver/base"
	"github.com/deis/builder/pkg/storage/driver/factory"
)

const (
	driverName       = "gcs"
	minChunkSize     = 256 * 1024
	defaultChunkSize = 20 * minChunkSize

	maxTries = 5
)

// driverParameters is a struct that encapsulates all of the driver parameters after all values have been set
type driverParameters struct {
	bucket        string
	config        *jwt.Config
	email         string
	privateKey    []byte
	clientOption  cloud.ClientOption
	rootDirectory string
	projectID     string
}

func init() {
	factory.Register(driverName, &gcsDriverFactory{})
}

// gcsDriverFactory implements the factory.StorageDriverFactory interface
type gcsDriverFactory struct{}

// Create StorageDriver from parameters
func (factory *gcsDriverFactory) Create(parameters map[string]string) (storagedriver.StorageDriver, error) {
	return FromParameters(parameters)
}

// driver is a storagedriver.StorageDriver implementation backed by GCS
// Objects are stored at absolute keys in the provided bucket.
type driver struct {
	client        *storage.Client
	bucket        string
	email         string
	privateKey    []byte
	rootDirectory string
	projectID     string
}

// FromParameters constructs a new Driver with a given parameters map
// Required parameters:
// - bucket
func FromParameters(parameters map[string]string) (storagedriver.StorageDriver, error) {
	bucket, ok := parameters["builder-bucket"]
	if !ok || fmt.Sprint(bucket) == "" {
		return nil, fmt.Errorf("No bucket parameter provided")
	}

	rootDirectory, ok := parameters["rootdirectory"]
	if !ok {
		rootDirectory = ""
	}

	var ts oauth2.TokenSource
	var key struct {
		ProjectID string `json:"project_id"`
	}
	jwtConf := new(jwt.Config)
	if keyfile, ok := parameters["keyfile"]; ok {
		jsonKey, err := ioutil.ReadFile(fmt.Sprint(keyfile))
		if err != nil {
			return nil, err
		}
		jwtConf, err = google.JWTConfigFromJSON(jsonKey, storage.ScopeFullControl)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(jsonKey, &key); err != nil {
			return nil, err
		}
		ts = jwtConf.TokenSource(context.Background())
	} else {
		var err error
		ts, err = google.DefaultTokenSource(context.Background(), storage.ScopeFullControl)
		if err != nil {
			return nil, err
		}

	}
	clientOption := cloud.WithTokenSource(ts)

	params := driverParameters{
		bucket:        fmt.Sprint(bucket),
		rootDirectory: fmt.Sprint(rootDirectory),
		email:         jwtConf.Email,
		privateKey:    jwtConf.PrivateKey,
		clientOption:  clientOption,
		projectID:     fmt.Sprint(key.ProjectID),
	}

	return New(params)
}

// New constructs a new driver
func New(params driverParameters) (storagedriver.StorageDriver, error) {
	rootDirectory := strings.Trim(params.rootDirectory, "/")
	if rootDirectory != "" {
		rootDirectory += "/"
	}

	adminClient, err := storage.NewAdminClient(context.Background(), params.projectID, params.clientOption)
	if err != nil {
		return nil, err
	}

	client, err := storage.NewClient(context.Background(), params.clientOption)
	if err != nil {
		return nil, err
	}

	if err = adminClient.CreateBucket(context.Background(), params.bucket, nil); err != nil {
		if e, ok := err.(*googleapi.Error); !ok || e.Code != 409 {
			fmt.Println(e.Code)
		}
	}

	d := &driver{
		bucket:        params.bucket,
		rootDirectory: rootDirectory,
		email:         params.email,
		privateKey:    params.privateKey,
		client:        client,
		projectID:     params.projectID,
	}

	return &base.Base{
		StorageDriver: d,
	}, nil
}

// Implement the storagedriver.StorageDriver interface

func (d *driver) Name() string {
	return driverName
}

func (d *driver) CheckConnectionStatus() (bool, error) {
	_, err := d.client.Bucket(d.bucket).List(context.Background(), nil)
	if err != nil {
		return false, err
	}
	return true, err
}

// GetContent retrieves the content stored at "path" as a []byte.
// This should primarily be used for small objects.
func (d *driver) GetContent(path string) ([]byte, error) {
	gcsContext := context.Background()
	name := d.pathToKey(path)
	var rc io.ReadCloser
	err := retry(func() error {
		var err error
		rc, err = d.client.Bucket(d.bucket).Object(name).NewReader(gcsContext)
		return err
	})
	if err == storage.ErrObjectNotExist {
		return nil, storagedriver.PathNotFoundError{Path: path}
	}
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	p, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// PutContent stores the []byte content at a location designated by "path".
// This should primarily be used for small objects.
func (d *driver) PutContent(path string, contents []byte) error {
	gcsContext := context.Background()
	name := d.pathToKey(path)
	return retry(func() error {
		wc := d.client.Bucket(d.bucket).Object(name).NewWriter(gcsContext)
		wc.ContentType = "application/octet-stream"
		return putContentsClose(wc, contents)
	})
}

// Stat retrieves the FileInfo for the given path, including the current
// size in bytes and the creation time.
func (d *driver) Stat(path string) (storagedriver.FileInfo, error) {
	var fi storagedriver.FileInfoFields
	//try to get as file
	gcsContext := context.Background()
	name := d.pathToKey(path)
	var obj *storage.ObjectAttrs
	err := retry(func() error {
		var err error
		obj, err = d.client.Bucket(d.bucket).Object(name).Attrs(gcsContext)
		return err
	})
	if err != nil {
		return nil, err
	}
	fi = storagedriver.FileInfoFields{
		Path:    path,
		Size:    obj.Size,
		ModTime: obj.Updated,
		IsDir:   false,
	}
	return storagedriver.FileInfoInternal{FileInfoFields: fi}, nil
}

func putContentsClose(wc *storage.Writer, contents []byte) error {
	size := len(contents)
	var nn int
	var err error
	for nn < size {
		n, err := wc.Write(contents[nn:size])
		nn += n
		if err != nil {
			break
		}
	}
	if err != nil {
		wc.CloseWithError(err)
		return err
	}
	return wc.Close()
}

type request func() error

func retry(req request) error {
	backoff := time.Second
	var err error
	for i := 0; i < maxTries; i++ {
		err = req()
		if err == nil {
			return nil
		}

		status, ok := err.(*googleapi.Error)
		if !ok || (status.Code != 429 && status.Code < http.StatusInternalServerError) {
			return err
		}

		time.Sleep(backoff - time.Second + (time.Duration(rand.Int31n(1000)) * time.Millisecond))
		if i <= 4 {
			backoff = backoff * 2
		}
	}
	return err
}

func (d *driver) pathToKey(path string) string {
	return strings.TrimRight(d.rootDirectory+strings.TrimLeft(path, "/"), "/")
}

func (d *driver) pathToDirKey(path string) string {
	return d.pathToKey(path) + "/"
}

func (d *driver) keyToPath(key string) string {
	return "/" + strings.Trim(strings.TrimPrefix(key, d.rootDirectory), "/")
}
