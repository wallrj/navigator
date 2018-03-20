package cassandra

import (
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type control struct {
	recorder record.EventRecorder
}

var _ Interface = &control{}

func NewControl(
	recorder record.EventRecorder,
) *control {
	return &control{
		recorder: recorder,
	}
}

func (e *control) Sync(c *v1alpha1.CassandraCluster) error {
	return nil
}
