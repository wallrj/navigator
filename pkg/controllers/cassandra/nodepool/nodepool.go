package nodepool

import (
	"k8s.io/api/apps/v1beta1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type defaultCassandraClusterNodepoolControl struct {
	kubeClient        kubernetes.Interface
	statefulSetLister appslisters.StatefulSetLister
	recorder          record.EventRecorder
}

var _ Interface = &defaultCassandraClusterNodepoolControl{}

func NewControl(
	kubeClient kubernetes.Interface,
	statefulSetLister appslisters.StatefulSetLister,
	recorder record.EventRecorder,
) Interface {
	return &defaultCassandraClusterNodepoolControl{
		kubeClient:        kubeClient,
		statefulSetLister: statefulSetLister,
		recorder:          recorder,
	}
}

func (e *defaultCassandraClusterNodepoolControl) clusterStatefulSets(
	cluster *v1alpha1.CassandraCluster,
) (results map[string]*v1beta1.StatefulSet, err error) {
	lister := e.statefulSetLister.StatefulSets(cluster.Namespace)
	selector, err := util.SelectorForCluster(cluster)
	if err != nil {
		return nil, err
	}
	existingSets, err := lister.List(selector)
	if err != nil {
		return nil, err
	}
	for _, set := range existingSets {
		err := util.OwnerCheck(set, cluster)
		if err != nil {
			continue
		}
		results[set.Name] = set
	}
	return results, nil
}

// Add a NodePoolStatus for each NodePool, only if a corresponding StatefulSet is found.
// Update the NodePoolStatus for each NodePool, using values from the corresponding StatefulSet.
// Remove the NodePoolStatus for NodePools that do not have a StatefulSet
// (the statefulset has been deleted unexpectedly)
// Remove the NodePoolStatus if there is no corresponding NodePool.
// (the statefulset has been removed by a DeleteNodePool action - not yet implemented)
func (e *defaultCassandraClusterNodepoolControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	if cluster.Status.NodePools == nil {
		cluster.Status.NodePools = map[string]v1alpha1.CassandraClusterNodePoolStatus{}
	}
	ssList, err := e.clusterStatefulSets(cluster)
	if err != nil {
		return err
	}
	nodePoolNames := sets.NewString()
	for _, np := range cluster.Spec.NodePools {
		nodePoolNames.Insert(np.Name)
		ss, setFound := ssList[np.Name]
		nps, npsFound := cluster.Status.NodePools[np.Name]
		if setFound {
			if !npsFound {
				cluster.Status.NodePools[np.Name] = nps
			}
			if nps.ReadyReplicas != ss.Status.ReadyReplicas {
				nps.ReadyReplicas = ss.Status.ReadyReplicas
				cluster.Status.NodePools[np.Name] = nps
			}
		} else {
			delete(cluster.Status.NodePools, np.Name)
		}
	}
	for npName := range cluster.Status.NodePools {
		if !nodePoolNames.Has(npName) {
			delete(cluster.Status.NodePools, npName)
		}
	}
	return nil
}
