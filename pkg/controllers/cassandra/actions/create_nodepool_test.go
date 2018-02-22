package actions_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/jetstack/navigator/internal/test/unit/framework"
	"github.com/jetstack/navigator/internal/test/util/generate"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/actions"
)

func TestCreateNodePool(t *testing.T) {
	type testT struct {
		kubeObjects         []runtime.Object
		navObjects          []runtime.Object
		cluster             generate.CassandraClusterConfig
		nodePool            generate.CassandraClusterNodePoolConfig
		expectedStatefulSet generate.StatefulSetConfig
		expectedErr         bool
	}
	tests := map[string]testT{
		"A statefulset is created if one does not already exist": {
			cluster: generate.CassandraClusterConfig{
				Name:      "cluster1",
				Namespace: "ns1",
			},
			nodePool: generate.CassandraClusterNodePoolConfig{
				Name: "pool1",
			},
			expectedStatefulSet: generate.StatefulSetConfig{
				Name:      "cass-cluster1-pool1",
				Namespace: "ns1",
				Replicas:  int32Ptr(0),
			},
		},
		"An error is returned if the statefulset already exists": {
			kubeObjects: []runtime.Object{
				generate.StatefulSet(
					generate.StatefulSetConfig{
						Name:      "cass-cluster1-pool1",
						Namespace: "ns1",
						Replicas:  int32Ptr(10),
					},
				),
			},
			cluster: generate.CassandraClusterConfig{Name: "cluster1", Namespace: "ns1"},
			nodePool: generate.CassandraClusterNodePoolConfig{
				Name: "pool1",
			},
			expectedStatefulSet: generate.StatefulSetConfig{
				Name:      "cass-cluster1-pool1",
				Namespace: "ns1",
				Replicas:  int32Ptr(10),
			},
			expectedErr: true,
		},
	}

	for name, test := range tests {
		t.Run(
			name,
			func(t *testing.T) {
				fixture := &framework.StateFixture{
					T:                t,
					KubeObjects:      test.kubeObjects,
					NavigatorObjects: test.navObjects,
				}
				fixture.Start()
				defer fixture.Stop()
				state := fixture.State()
				a := &actions.CreateNodePool{
					Cluster:  generate.CassandraCluster(test.cluster),
					NodePool: generate.CassandraClusterNodePool(test.nodePool),
				}
				err := a.Execute(state)
				if !test.expectedErr && err != nil {
					t.Errorf("Unexpected error: %s", err)
				}
				if test.expectedErr && err == nil {
					t.Errorf("Expected an error")
				}
				actualStatefulSet, err := fixture.KubeClient().
					AppsV1beta1().
					StatefulSets(test.expectedStatefulSet.Namespace).
					Get(test.expectedStatefulSet.Name, metav1.GetOptions{})
				if err != nil {
					t.Fatalf("Unexpected error retrieving statefulset: %v", err)
				}
				generate.AssertStatefulSetMatches(t, test.expectedStatefulSet, actualStatefulSet)
			},
		)
	}
}
