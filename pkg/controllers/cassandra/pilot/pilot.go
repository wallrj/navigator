package pilot

import (
	"k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/pkg/errors"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	navigator "github.com/jetstack/navigator/pkg/client/clientset/versioned"
	navlisters "github.com/jetstack/navigator/pkg/client/listers/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
)

const (
	HashAnnotationKey = "navigator.jetstack.io/cassandra-pilot-hash"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type pilotControl struct {
	naviClient   navigator.Interface
	pilots       navlisters.PilotLister
	pods         corelisters.PodLister
	statefulSets appslisters.StatefulSetLister
	recorder     record.EventRecorder
}

var _ Interface = &pilotControl{}

func NewControl(
	naviClient navigator.Interface,
	pilots navlisters.PilotLister,
	pods corelisters.PodLister,
	statefulSets appslisters.StatefulSetLister,
	recorder record.EventRecorder,
) *pilotControl {
	return &pilotControl{
		naviClient:   naviClient,
		pilots:       pilots,
		pods:         pods,
		statefulSets: statefulSets,
		recorder:     recorder,
	}

}

func (c *pilotControl) ensurePilot(cluster *v1alpha1.CassandraCluster, pod *v1.Pod) error {
	desiredPilot := PilotForCluster(cluster, pod)
	existingPilot, err := c.pilots.Pilots(desiredPilot.Namespace).Get(desiredPilot.Name)
	// If Pilot already exists, check that it belongs to this cluster
	if err == nil {
		return errors.Wrap(
			util.OwnerCheck(existingPilot, cluster),
			"owner check error",
		)
	}
	// The only error we expect is that the pilot does not exist.
	if !k8sErrors.IsNotFound(err) {
		return errors.Wrap(err, "unable to get pilot")
	}
	_, err = c.naviClient.NavigatorV1alpha1().Pilots(desiredPilot.Namespace).Create(desiredPilot)
	if err == nil || k8sErrors.IsAlreadyExists(err) {
		return nil
	}
	return errors.Wrap(err, "unable to create pilot")
}

// Create a Pilot for every pod that has a matching ClusterName label and a NodePoolNameLabelKey
// Delete Pilots only if there is no corresponding pod and if the index is higher than the current replica count.
func (c *pilotControl) Sync(cluster *v1alpha1.CassandraCluster) error {
	selector, err := util.SelectorForClusterNodePools(cluster)
	if err != nil {
		return errors.Wrap(err, "unable to create cluster nodepools selector")
	}
	pods, err := c.pods.Pods(cluster.Namespace).List(selector)
	if err != nil {
		return errors.Wrap(err, "unable to list pods")
	}
	for _, pod := range pods {
		err := c.ensurePilot(cluster, pod)
		if err != nil {
			return errors.Wrap(err, "unable to ensure pilot")
		}
	}
	return nil
}

func PilotForCluster(cluster *v1alpha1.CassandraCluster, pod *v1.Pod) *v1alpha1.Pilot {
	return &v1alpha1.Pilot{
		ObjectMeta: metav1.ObjectMeta{
			Name:            pod.Name,
			Namespace:       pod.Namespace,
			Labels:          util.ClusterLabels(cluster),
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(cluster)},
		},
	}
}
