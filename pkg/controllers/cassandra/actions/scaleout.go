package actions

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
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
	ss = ss.DeepCopy()
	if ss.Spec.Replicas == nil || *ss.Spec.Replicas < a.NodePool.Replicas {
		ss.Spec.Replicas = &a.NodePool.Replicas
		_, err = s.Clientset.AppsV1beta1().StatefulSets(ss.Namespace).Update(ss)
		if err == nil {
			s.Recorder.Eventf(
				a.Cluster,
				corev1.EventTypeNormal,
				a.Name(),
				"Scaled node pool %q to %d replicas", a.NodePool.Name, a.NodePool.Replicas,
			)
		}
		return err
	}
	if *ss.Spec.Replicas > a.NodePool.Replicas {
		return fmt.Errorf(
			"the NodePool.Replicas value (%d) "+
				"is less than the existing StatefulSet.Replicas value (%d)",
			a.NodePool.Replicas, *ss.Spec.Replicas,
		)
	}
	return nil
}
