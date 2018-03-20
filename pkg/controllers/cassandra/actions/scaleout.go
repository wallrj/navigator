package actions

import (
	"fmt"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
)

type ScaleOut struct {
	Cluster  *v1alpha1.CassandraCluster
	NodePool *v1alpha1.CassandraClusterNodePool
}

var _ controllers.Action = &ScaleOut{}

func (a *ScaleOut) Name() string {
	return "ScaleOut"
}

func (a *ScaleOut) Execute(s *controllers.State) error {
	ss := nodepool.StatefulSetForCluster(a.Cluster, a.NodePool)
	ss, err := s.StatefulSetLister.StatefulSets(ss.Namespace).Get(ss.Name)
	if err != nil {
		return err
	}
	switch {
	case *ss.Spec.Replicas < a.NodePool.Replicas:
		ss = ss.DeepCopy()
		ss.Spec.Replicas = util.Int32Ptr(*ss.Spec.Replicas + 1)
		_, err = s.Clientset.AppsV1beta1().StatefulSets(ss.Namespace).Update(ss)
		if err != nil {
			return err
		}
	case *ss.Spec.Replicas > a.NodePool.Replicas:
		return fmt.Errorf(
			"the NodePool.Replicas value (%d) "+
				"is less than the existing StatefulSet.Replicas value (%d)",
			a.NodePool.Replicas, *ss.Spec.Replicas,
		)
	}
	return nil
}
