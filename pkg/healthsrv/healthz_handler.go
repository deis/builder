package healthsrv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	s3 "github.com/aws/aws-sdk-go/service/s3"
)

type healthZRespBucket struct {
	Name         string    `json:"name"`
	CreationDate time.Time `json:"creation_date"`
}

func convertBucket(b *s3.Bucket) healthZRespBucket {
	return healthZRespBucket{
		Name:         *b.Name,
		CreationDate: *b.CreationDate,
	}
}

type healthZResp struct {
	Buckets []healthZRespBucket `json:"buckets"`
}

func healthZHandler(s3Client *s3.S3) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lbOut, err := s3Client.ListBuckets(s3.ListBucketsInput{})
		if err != nil {
			str := fmt.Sprintf("Error listing buckets (%s)", err)
			log.Printf(str)
			http.Error(w, str, http.StatusInternalServerError)
			return
		}
		var rsp healthZResp
		for _, buck := range lbOut.Buckets {
			rsp.Buckets = append(rsp.Buckets, convertBucket(buck))
		}
		if err := json.NewEncoder(w).Encode(rsp); err != nil {
			str := fmt.Sprintf("Error encoding JSON (%s)", err)
			http.Error(w, str, http.StatusInternalServerError)
			return
		}
		// TODO: check k8s API
		// TODO: check server is running
	})
}
