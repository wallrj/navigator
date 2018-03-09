package nodes

import (
	"fmt"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"

	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

type Interface interface {
	Sync(*v1alpha1.CassandraCluster) error
}

type control struct {
	kubeClient    kubernetes.Interface
	serviceLister corelisters.ServiceLister
	recorder      record.EventRecorder
}

var _ Interface = &control{}

func NewControl(
	kubeClient kubernetes.Interface,
	serviceLister corelisters.ServiceLister,
	recorder record.EventRecorder,
) Interface {
	return &control{
		kubeClient:    kubeClient,
		serviceLister: serviceLister,
		recorder:      recorder,
	}
}

func (c *control) Sync(cluster *v1alpha1.CassandraCluster) error {
	service := ServiceForCluster(cluster)
	_, err := c.serviceLister.Services(service.Namespace).Get(service.Name)
	if err == nil {
		glog.V(4).Infof("Service already exists: %s", service.Name)
		return nil
	}
	if !k8sErrors.IsNotFound(err) {
		return err
	}
	glog.V(4).Infof("Creating service: %s", service.Name)
	_, err = c.kubeClient.CoreV1().Services(service.Namespace).Create(service)
	if k8sErrors.IsAlreadyExists(err) {
		glog.V(4).Infof("Service exists: %s", service.Name)
		return nil
	}
	return err
}

func ServiceForCluster(
	cluster *v1alpha1.CassandraCluster,
) *apiv1.Service {
	clusterLabels := util.ClusterLabels(cluster)
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            fmt.Sprintf("%s-nodes", util.ResourceBaseName(cluster)),
			Namespace:       cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{util.NewControllerRef(cluster)},
			Labels:          clusterLabels,
		},
		Spec: apiv1.ServiceSpec{
			ClusterIP: "None",
			Type:      apiv1.ServiceTypeClusterIP,
			Selector:  clusterLabels,
			// Headless service should not require a port.
			// But without it, DNS records are not registered.
			// See https://github.com/kubernetes/kubernetes/issues/55158
			Ports: []apiv1.ServicePort{{Port: 65535}},
		},
	}
}
