package gitreceive

import (
	"errors"

	"github.com/coreos/go-etcd/etcd"
)

const (
	builderKey = "/deis/controller/builderKey"
)

var (
	errNoNode = errors.New("no etcd node")
)

func getBuilderKey(etcdClient *etcd.Client) (string, error) {
	resp, err := etcdClient.Get(builderKey, false, false)
	if err != nil {
		return "", err
	}
	if resp.Node == nil {
		return "", errNoNode
	}
	return resp.Node.Value, nil
}
