package generate

import (
	"testing"

	"github.com/coreos/go-semver/semver"
	apps "k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
)

type PilotConfig struct {
	Name, Namespace   string
	Cluster, NodePool string
	Documents         *int64
	Version           string
}

func Pilot(c PilotConfig) *v1alpha1.Pilot {
	if c.Namespace == "" {
		c.Namespace = "default"
	}
	labels := map[string]string{}
	labels[v1alpha1.ElasticsearchClusterNameLabel] = c.Cluster
	labels[v1alpha1.ElasticsearchNodePoolNameLabel] = c.NodePool
	var version semver.Version
	if c.Version != "" {
		version = *semver.New(c.Version)
	}
	return &v1alpha1.Pilot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
			Labels:    labels,
		},
		Status: v1alpha1.PilotStatus{
			Elasticsearch: &v1alpha1.ElasticsearchPilotStatus{
				Documents: c.Documents,
				Version:   version,
			},
		},
	}
}

type ClusterConfig struct {
	Name, Namespace string
	NodePools       []v1alpha1.ElasticsearchClusterNodePool
	Version         string
	ClusterConfig   v1alpha1.NavigatorClusterConfig
	Health          v1alpha1.ElasticsearchClusterHealth
}

func Cluster(c ClusterConfig) *v1alpha1.ElasticsearchCluster {
	if c.Namespace == "" {
		c.Namespace = "default"
	}
	return &v1alpha1.ElasticsearchCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
		},
		Spec: v1alpha1.ElasticsearchClusterSpec{
			NavigatorClusterConfig: c.ClusterConfig,
			Version:                *semver.New(c.Version),
			NodePools:              c.NodePools,
		},
		Status: v1alpha1.ElasticsearchClusterStatus{
			Health: c.Health,
		},
	}
}

type StatefulSetConfig struct {
	Name, Namespace                  string
	Replicas                         *int32
	Version, Image                   string
	Partition                        *int32
	CurrentRevision, UpdateRevision  string
	CurrentReplicas, UpdatedReplicas int32
	ReadyReplicas                    int32
}

func StatefulSet(c StatefulSetConfig) *apps.StatefulSet {
	if c.Namespace == "" {
		c.Namespace = "default"
	}
	return &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
			Annotations: map[string]string{
				v1alpha1.ElasticsearchNodePoolVersionAnnotation: c.Version,
			},
		},
		Spec: apps.StatefulSetSpec{
			Replicas: c.Replicas,
			UpdateStrategy: apps.StatefulSetUpdateStrategy{
				RollingUpdate: &apps.RollingUpdateStatefulSetStrategy{
					Partition: c.Partition,
				},
			},
			Template: core.PodTemplateSpec{
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Image: c.Image,
						},
					},
				},
			},
		},
		Status: apps.StatefulSetStatus{
			CurrentRevision: c.CurrentRevision,
			UpdateRevision:  c.UpdateRevision,
			CurrentReplicas: c.CurrentReplicas,
			UpdatedReplicas: c.UpdatedReplicas,
			ReadyReplicas:   c.ReadyReplicas,
		},
	}
}

func AssertStatefulSetMatches(t *testing.T, expected StatefulSetConfig, actual *apps.StatefulSet) {
	if actual.Name != expected.Name {
		t.Errorf("Name %q != %q", expected.Name, actual.Name)
	}
	if actual.Namespace != expected.Namespace {
		t.Errorf("Namespace %q != %q", expected.Namespace, actual.Namespace)
	}
	if expected.Replicas != nil {
		if actual.Spec.Replicas == nil {
			t.Errorf("Replicas %d != %v", *expected.Replicas, nil)
		} else {
			if *actual.Spec.Replicas != *expected.Replicas {
				t.Errorf("Replicas %d != %d", *expected.Replicas, *actual.Spec.Replicas)
			}
		}
	}
}

type CassandraClusterConfig struct {
	Name, Namespace string
}

func CassandraCluster(c CassandraClusterConfig) *v1alpha1.CassandraCluster {
	return &v1alpha1.CassandraCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.Name,
			Namespace: c.Namespace,
		},
	}
}

type CassandraClusterNodePoolConfig struct {
	Name     string
	Replicas int32
}

func CassandraClusterNodePool(c CassandraClusterNodePoolConfig) *v1alpha1.CassandraClusterNodePool {
	return &v1alpha1.CassandraClusterNodePool{
		Name:     c.Name,
		Replicas: c.Replicas,
	}
}
