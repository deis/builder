package healthsrv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	s3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/deis/builder/pkg/gitreceive/log"
	"github.com/deis/builder/pkg/sshd"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
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
	Namespaces       []string            `json:"k8s_namespaces"`
	S3Buckets        []healthZRespBucket `json:"s3_buckets"`
	SSHServerStarted bool                `json:"ssh_server_started"`
}

func healthZHandler(nsLister NamespaceLister, bLister BucketLister, serverCircuit *sshd.Circuit) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// there's a race between the boolean eval and the HTTP error returned, but k8s will repeat the health probe request
		// and effectively re-evaluate the boolean.
		if serverCircuit.State() != sshd.ClosedState {
			str := fmt.Sprintf("SSH Server is not yet started")
			log.Err(str)
			http.Error(w, str, http.StatusAccepted)
			return
		}
		lbOut, err := bLister.ListBuckets(&s3.ListBucketsInput{})
		if err != nil {
			str := fmt.Sprintf("Error listing buckets (%s)", err)
			log.Err(str)
			http.Error(w, str, http.StatusInternalServerError)
			return
		}
		var rsp healthZResp
		for _, buck := range lbOut.Buckets {
			rsp.S3Buckets = append(rsp.S3Buckets, convertBucket(buck))
		}

		nsList, err := nsLister.List(labels.Everything(), fields.Everything())
		if err != nil {
			str := fmt.Sprintf("Error listing buckets (%s)", err)
			log.Err(str)
			http.Error(w, str, http.StatusInternalServerError)
			return
		}
		for _, ns := range nsList.Items {
			rsp.Namespaces = append(rsp.Namespaces, ns.Name)
		}
		rsp.SSHServerStarted = true
		if err := json.NewEncoder(w).Encode(rsp); err != nil {
			str := fmt.Sprintf("Error encoding JSON (%s)", err)
			http.Error(w, str, http.StatusInternalServerError)
			return
		}
		// TODO: check server is running
	})
}
