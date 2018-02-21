package actions

import (
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	// k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

type CreateNodePool struct {
	Cluster  *v1alpha1.CassandraCluster
	NodePool *v1alpha1.CassandraClusterNodePool
}

var _ controllers.Action = &CreateNodePool{}

func (a *CreateNodePool) Name() string {
	return "CreateNodePool"
}

func (a *CreateNodePool) Execute(s *controllers.State) error {
	ss := nodepool.StatefulSetForCluster(a.Cluster, a.NodePool)
	_, err := s.Clientset.AppsV1beta1().StatefulSets(ss.Namespace).Create(ss)
	// XXX: Should this be idempotent?
	// if k8sErrors.IsAlreadyExists(err) {
	//	return nil
	// }
	return err
}
