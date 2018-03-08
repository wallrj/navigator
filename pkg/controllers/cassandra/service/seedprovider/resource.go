package seedprovider

import (
	apiv1 "k8s.io/api/core/v1"

	v1alpha1 "github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	serviceutil "github.com/jetstack/navigator/pkg/controllers/cassandra/service/util"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
)

const (
	SeedLabelKey   = "navigator.jetstack.io/cassandra-seed"
	SeedLabelValue = "true"
)

func ServiceForCluster(
	cluster *v1alpha1.CassandraCluster,
) *apiv1.Service {
	return updateServiceForCluster(cluster, &apiv1.Service{})
}

func updateServiceForCluster(
	cluster *v1alpha1.CassandraCluster,
	service *apiv1.Service,
) *apiv1.Service {
	service = service.DeepCopy()
	service = serviceutil.SetStandardServiceAttributes(cluster, service)
	service.SetName(util.SeedProviderServiceName(cluster))
	service.Spec.Type = apiv1.ServiceTypeClusterIP
	service.Spec.ClusterIP = "None"

	// Only mark nodes explicitly labeled as seeds as seed nodes
	service.Spec.Selector[SeedLabelKey] = SeedLabelValue

	// Headless service should not require a port.
	// But without it, DNS records are not registered.
	// See https://github.com/kubernetes/kubernetes/issues/55158
	service.Spec.Ports = []apiv1.ServicePort{
		{
			Port: 65535,
		},
	}
	return service
}
