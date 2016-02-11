package healthsrv

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	s3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/deis/builder/pkg/sshd"
	"github.com/deis/pkg/log"
	"k8s.io/kubernetes/pkg/api"
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

func marshalHealthZResp(w http.ResponseWriter, rsp healthZResp) {
	if err := json.NewEncoder(w).Encode(rsp); err != nil {
		str := fmt.Sprintf("Error encoding JSON (%s)", err)
		http.Error(w, str, http.StatusInternalServerError)
		return
	}
}

func healthZHandler(nsLister NamespaceLister, bLister BucketLister, serverCircuit *sshd.Circuit) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stopCh := make(chan struct{})

		serverStateCh := make(chan struct{})
		serverStateErrCh := make(chan error)
		go circuitState(serverCircuit, serverStateCh, serverStateErrCh, stopCh)

		listBucketsCh := make(chan *s3.ListBucketsOutput)
		listBucketsErrCh := make(chan error)
		go listBuckets(bLister, listBucketsCh, listBucketsErrCh, stopCh)

		namespaceListerCh := make(chan *api.NamespaceList)
		namespaceListerErrCh := make(chan error)
		go listNamespaces(nsLister, namespaceListerCh, namespaceListerErrCh, stopCh)

		var rsp healthZResp
		serverState, bucketState, namespaceState := false, false, false
		for {
			select {
			case err := <-serverStateErrCh:
				log.Err(err.Error())
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				close(stopCh)
				return
			case err := <-listBucketsErrCh:
				str := fmt.Sprintf("Error listing buckets (%s)", err)
				log.Err(str)
				http.Error(w, str, http.StatusServiceUnavailable)
				close(stopCh)
				return
			case err := <-namespaceListerErrCh:
				str := fmt.Sprintf("Error listing namespaces (%s)", err)
				log.Err(str)
				http.Error(w, str, http.StatusServiceUnavailable)
				close(stopCh)
				return
			case <-serverStateCh:
				serverState = true
				rsp.SSHServerStarted = true
				if serverState && bucketState && namespaceState {
					marshalHealthZResp(w, rsp)
					return
				}
			case lbOut := <-listBucketsCh:
				bucketState = true
				for _, buck := range lbOut.Buckets {
					rsp.S3Buckets = append(rsp.S3Buckets, convertBucket(buck))
				}
				if serverState && bucketState && namespaceState {
					marshalHealthZResp(w, rsp)
					return
				}
			case nsList := <-namespaceListerCh:
				namespaceState = true
				for _, ns := range nsList.Items {
					rsp.Namespaces = append(rsp.Namespaces, ns.Name)
				}
				if serverState && bucketState && namespaceState {
					marshalHealthZResp(w, rsp)
					return
				}
			}
		}
	})
}
