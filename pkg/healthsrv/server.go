package healthsrv

import (
	"fmt"
	"net/http"

	s3 "github.com/aws/aws-sdk-go/service/s3"
)

const (
	DefaultHost = "0.0.0.0"
	DefaultPort = 8082
)

// Start starts the healthcheck server on $host:$port and blocks. It only returns if the server fails, with the indicative error
func Start(port int, s3Client *s3.S3) error {
	mux := http.NewServeMux()
	mux.Handle("/healthz", healthZHandler(s3Client))

	hostStr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(hostStr, mux)
}
