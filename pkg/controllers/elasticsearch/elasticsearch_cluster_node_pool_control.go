package elasticsearch

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1beta1"
	extensionslisters "k8s.io/client-go/listers/extensions/v1beta1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/record"

	"gitlab.jetstack.net/marshal/colonel/pkg/api/v1"
)

type ElasticsearchClusterNodePoolControl interface {
	CreateElasticsearchClusterNodePool(*v1.ElasticsearchCluster, *v1.ElasticsearchClusterNodePool) error
	UpdateElasticsearchClusterNodePool(*v1.ElasticsearchCluster, *v1.ElasticsearchClusterNodePool) error
	DeleteElasticsearchClusterNodePool(*v1.ElasticsearchCluster, *v1.ElasticsearchClusterNodePool) error
}

type defaultElasticsearchClusterNodePoolControl struct {
	kubeClient        *kubernetes.Clientset
	statefulSetLister appslisters.StatefulSetLister
	deploymentLister  extensionslisters.DeploymentLister

	recorder record.EventRecorder
}

type statefulElasticsearchClusterNodePoolControl struct {
	kubeClient        *kubernetes.Clientset
	statefulSetLister appslisters.StatefulSetLister
	deploymentLister  extensionslisters.DeploymentLister

	recorder record.EventRecorder
}

var _ ElasticsearchClusterNodePoolControl = &defaultElasticsearchClusterNodePoolControl{}
var _ ElasticsearchClusterNodePoolControl = &statefulElasticsearchClusterNodePoolControl{}

func NewElasticsearchClusterNodePoolControl(
	kubeClient *kubernetes.Clientset,
	statefulSetLister appslisters.StatefulSetLister,
	deploymentLister extensionslisters.DeploymentLister,
	recorder record.EventRecorder,
) ElasticsearchClusterNodePoolControl {
	return &defaultElasticsearchClusterNodePoolControl{
		kubeClient:        kubeClient,
		statefulSetLister: statefulSetLister,
		deploymentLister:  deploymentLister,
		recorder:          recorder,
	}
}

func NewStatefulElasticsearchClusterNodePoolControl(
	kubeClient *kubernetes.Clientset,
	statefulSetLister appslisters.StatefulSetLister,
	deploymentLister extensionslisters.DeploymentLister,
	recorder record.EventRecorder,
) ElasticsearchClusterNodePoolControl {
	return &statefulElasticsearchClusterNodePoolControl{
		kubeClient:        kubeClient,
		statefulSetLister: statefulSetLister,
		deploymentLister:  deploymentLister,
		recorder:          recorder,
	}
}

func (e *defaultElasticsearchClusterNodePoolControl) CreateElasticsearchClusterNodePool(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) error {
	depl, err := nodePoolDeployment(c, np)

	if err != nil {
		return fmt.Errorf("error generating deployment manifest: %s", err.Error())
	}

	depl, err = e.kubeClient.Extensions().Deployments(c.Namespace).Create(depl)

	if err != nil {
		return fmt.Errorf("error creating deployment: %s", err.Error())
	}

	return nil
}

func (e *defaultElasticsearchClusterNodePoolControl) UpdateElasticsearchClusterNodePool(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) error {
	depl, err := nodePoolDeployment(c, np)

	if err != nil {
		return fmt.Errorf("error generating deployment manifest: %s", err.Error())
	}

	depl, err = e.kubeClient.Extensions().Deployments(c.Namespace).Update(depl)

	if err != nil {
		return fmt.Errorf("error creating deployment: %s", err.Error())
	}

	return nil
}

func (e *defaultElasticsearchClusterNodePoolControl) DeleteElasticsearchClusterNodePool(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) error {
	err := e.kubeClient.Extensions().Deployments(c.Namespace).Delete(nodePoolResourceName(c, np), &metav1.DeleteOptions{OrphanDependents: &falseVar})

	if err != nil {
		return fmt.Errorf("error deleting deployment: %s", err.Error())
	}

	return nil
}

func (e *statefulElasticsearchClusterNodePoolControl) CreateElasticsearchClusterNodePool(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) error {
	ss, err := nodePoolStatefulSet(c, np)

	if err != nil {
		return fmt.Errorf("error generating statefulset manifest: %s", err.Error())
	}

	ss, err = e.kubeClient.Apps().StatefulSets(c.Namespace).Create(ss)

	if err != nil {
		return fmt.Errorf("error creating statefulset: %s", err.Error())
	}

	return nil
}

func (e *statefulElasticsearchClusterNodePoolControl) UpdateElasticsearchClusterNodePool(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) error {
	ss, err := nodePoolStatefulSet(c, np)

	if err != nil {
		return fmt.Errorf("error generating statefulset manifest: %s", err.Error())
	}

	ss, err = e.kubeClient.Apps().StatefulSets(c.Namespace).Update(ss)

	if err != nil {
		return fmt.Errorf("error updating statefulset: %s", err.Error())
	}

	return nil
}

func (e *statefulElasticsearchClusterNodePoolControl) DeleteElasticsearchClusterNodePool(c *v1.ElasticsearchCluster, np *v1.ElasticsearchClusterNodePool) error {
	err := e.kubeClient.Apps().StatefulSets(c.Namespace).Delete(nodePoolResourceName(c, np), &metav1.DeleteOptions{OrphanDependents: &falseVar})

	if err != nil {
		return fmt.Errorf("error deleting statefulset: %s", err.Error())
	}

	return nil
}

// recordNodePoolEvent records an event for verb applied to a NodePool in an ElasticsearchCluster. If err is nil the generated event will
// have a reason of v1.EventTypeNormal. If err is not nil the generated event will have a reason of v1.EventTypeWarning.
func (e *defaultElasticsearchClusterNodePoolControl) recordNodePoolEvent(verb string, cluster *v1.ElasticsearchCluster, pool *v1.ElasticsearchClusterNodePool, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s NodePool %s in ElasticsearchCluster %s successful",
			strings.ToLower(verb), pool.Name, cluster.Name)
		e.recorder.Event(cluster, apiv1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s NodePool %s in ElasticsearchCluster %s failed error: %s",
			strings.ToLower(verb), pool.Name, cluster.Name, err)
		e.recorder.Event(cluster, apiv1.EventTypeWarning, reason, message)
	}
}

// recordNodePoolEvent records an event for verb applied to a NodePool in an ElasticsearchCluster. If err is nil the generated event will
// have a reason of v1.EventTypeNormal. If err is not nil the generated event will have a reason of v1.EventTypeWarning.
func (e *statefulElasticsearchClusterNodePoolControl) recordNodePoolEvent(verb string, cluster *v1.ElasticsearchCluster, pool *v1.ElasticsearchClusterNodePool, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s StatefulNodePool %s in ElasticsearchCluster %s successful",
			strings.ToLower(verb), pool.Name, cluster.Name)
		e.recorder.Event(cluster, apiv1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s StatefulNodePool %s in ElasticsearchCluster %s failed error: %s",
			strings.ToLower(verb), pool.Name, cluster.Name, err)
		e.recorder.Event(cluster, apiv1.EventTypeWarning, reason, message)
	}
}
