package pilot

import (
	"strconv"
	"strings"

	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	"github.com/golang/glog"
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

	sets, err := c.statefulSets.StatefulSets(cluster.Namespace).List(selector)
	if err != nil {
		return errors.Wrap(err, "unable to list statefulsets")
	}

	setsByNodePoolName := map[string]*v1beta1.StatefulSet{}
	for _, set := range sets {
		setNodePoolName := set.Labels[v1alpha1.CassandraNodePoolNameLabel]
		setsByNodePoolName[setNodePoolName] = set
	}

	pilots, err := c.pilots.Pilots(cluster.Namespace).List(selector)
	if err != nil {
		return errors.Wrap(err, "unable to list pilots")
	}

	for _, pilot := range pilots {
		pilotNodePoolName := pilot.Labels[v1alpha1.CassandraNodePoolNameLabel]
		setForPilot := setsByNodePoolName[pilotNodePoolName]
		pilotIndexLabel := pilot.Labels[v1alpha1.CassandraNodePoolIndexLabel]
		pilotIndex, err := strconv.Atoi(pilotIndexLabel)
		if err != nil {
			glog.Errorf(
				"Unable to parse pilot %s/%s index: %q",
				pilot.Namespace, pilot.Name, pilotIndexLabel,
			)
		}
		if int32(pilotIndex) > setForPilot.Status.CurrentReplicas {
			err := c.naviClient.NavigatorV1alpha1().
				Pilots(cluster.Namespace).Delete(pilot.Name, &metav1.DeleteOptions{})
			if err != nil {
				if !k8sErrors.IsNotFound(err) {
					return errors.Wrapf(
						err, "unable to delete pilot %s/%s", pilot.Namespace, pilot.Name,
					)
				}
			}
		}
	}
	return nil
}

func parsePodIndex(podName string) (int, bool) {
	parts := strings.Split(podName, "-")
	if len(parts) < 2 {
		return -1, false
	}
	index, err := strconv.Atoi(parts[len(parts)])
	if err != nil {
		return -1, false
	}
	return index, true
}

func PilotForCluster(cluster *v1alpha1.CassandraCluster, pod *v1.Pod) *v1alpha1.Pilot {
	o := &v1alpha1.Pilot{
		ObjectMeta: metav1.ObjectMeta{
			Name:            pod.Name,
			Namespace:       pod.Namespace,
			Labels:          util.ClusterLabels(cluster),
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(cluster)},
		},
	}
	o.Labels[v1alpha1.CassandraNodePoolNameLabel] = pod.Labels[v1alpha1.CassandraNodePoolNameLabel]
	index, found := parsePodIndex(pod.Name)
	if found {
		o.Labels[v1alpha1.CassandraNodePoolIndexLabel] = strconv.Itoa(index)
	}
	return o
}
