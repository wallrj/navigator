package pilot_test

import (
	"fmt"
	"testing"

	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kr/pretty"

	"github.com/jetstack/navigator/internal/test/unit/framework"
	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/pilot"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
)

func clusterPod(cluster *v1alpha1.CassandraCluster, set *v1beta1.StatefulSet, index int32) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", set.Name, index),
			Namespace: set.Namespace,
			Labels: map[string]string{
				v1alpha1.CassandraClusterNameLabel:  cluster.Name,
				v1alpha1.CassandraNodePoolNameLabel: set.Labels[v1alpha1.CassandraNodePoolNameLabel],
			},
		},
	}
}

func TestPilotSync(t *testing.T) {
	cluster1 := casstesting.ClusterForTest()
	cluster1np1 := &cluster1.Spec.NodePools[0]
	cluster1np1set := nodepool.StatefulSetForCluster(cluster1, cluster1np1)
	cluster1np1set.Status.CurrentReplicas = 1
	cluster1np1pod0 := clusterPod(cluster1, cluster1np1set, 0)
	cluster1np1pod1 := clusterPod(cluster1, cluster1np1set, 1)
	cluster1np1pilot0 := pilot.PilotForCluster(cluster1, cluster1np1pod0)
	cluster1np1pilot1 := pilot.PilotForCluster(cluster1, cluster1np1pod1)
	cluster1np1pilot0foreign := cluster1np1pilot0.DeepCopy()
	cluster1np1pilot0foreign.SetOwnerReferences([]metav1.OwnerReference{})

	cluster2 := casstesting.ClusterForTest()
	cluster2.SetName("cluster2")
	cluster2.SetUID("uid2")
	cluster2np1 := &cluster2.Spec.NodePools[0]
	cluster2np1set := nodepool.StatefulSetForCluster(cluster2, cluster2np1)
	cluster2np1pod0 := clusterPod(cluster2, cluster2np1set, 0)

	type testT struct {
		kubeObjects []runtime.Object
		navObjects  []runtime.Object
		cluster     *v1alpha1.CassandraCluster
		assertions  func(*testing.T, *controllers.State)
		expectErr   bool
	}

	tests := map[string]testT{
		"each cluster pod gets a pilot": {
			kubeObjects: []runtime.Object{
				cluster1np1set,
				cluster1np1pod0,
				cluster1np1pod1,
				cluster2np1set,
				cluster2np1pod0,
			},
			cluster: cluster1,
			assertions: func(t *testing.T, state *controllers.State) {
				pilots, err := state.NavigatorClientset.
					Navigator().Pilots(cluster1.Namespace).List(metav1.ListOptions{})
				if err != nil {
					t.Fatal(err)
				}
				expectedPilotCount := 2
				pilotCount := len(pilots.Items)
				if pilotCount != expectedPilotCount {
					t.Log(pretty.Sprint(pilots))
					t.Errorf("Unexpected pilot count: %d != %d", expectedPilotCount, pilotCount)
				}
			},
		},
		"non-cluster pods are ignored": {
			kubeObjects: []runtime.Object{
				cluster1np1set,
				cluster1np1pod0,
				cluster2np1set,
				cluster2np1pod0,
			},
			cluster: cluster1,
			assertions: func(t *testing.T, state *controllers.State) {
				pilots, err := state.NavigatorClientset.
					Navigator().Pilots(cluster1.Namespace).List(metav1.ListOptions{})
				if err != nil {
					t.Fatal(err)
				}
				expectedPilotCount := 1
				pilotCount := len(pilots.Items)
				if pilotCount != expectedPilotCount {
					t.Log(pretty.Sprint(pilots))
					t.Errorf("Unexpected pilot count: %d != %d", expectedPilotCount, pilotCount)
				}
			},
		},
		"no error if pilot exists": {
			kubeObjects: []runtime.Object{
				cluster1np1set,
				cluster1np1pod0,
			},
			navObjects: []runtime.Object{cluster1np1pilot0},
			cluster:    cluster1,
		},
		"error if foreign owned": {
			kubeObjects: []runtime.Object{cluster1np1pod0},
			navObjects:  []runtime.Object{cluster1np1pilot0foreign},
			cluster:     cluster1,
			expectErr:   true,
		},
		"do not delete if pod exists": {
			kubeObjects: []runtime.Object{
				cluster1np1set,
				cluster1np1pod0,
				cluster1np1pod1,
			},
			navObjects: []runtime.Object{
				cluster1np1pilot0,
				cluster1np1pilot1,
			},
			cluster: cluster1,
			assertions: func(t *testing.T, state *controllers.State) {
				pilots, err := state.NavigatorClientset.
					Navigator().Pilots(cluster1.Namespace).List(metav1.ListOptions{})
				if err != nil {
					t.Fatal(err)
				}
				expectedPilotCount := 2
				pilotCount := len(pilots.Items)
				if pilotCount != expectedPilotCount {
					t.Log(pretty.Sprint(pilots))
					t.Errorf("Unexpected pilot count: %d != %d", expectedPilotCount, pilotCount)
				}
			},
		},
		"pilot deleted": {
			kubeObjects: []runtime.Object{
				cluster1np1set,
				cluster1np1pod0,
			},
			navObjects: []runtime.Object{
				cluster1np1pilot0,
				cluster1np1pilot1,
			},
			cluster: cluster1,
			assertions: func(t *testing.T, state *controllers.State) {
				pilots, err := state.NavigatorClientset.
					Navigator().Pilots(cluster1.Namespace).List(metav1.ListOptions{})
				if err != nil {
					t.Fatal(err)
				}
				expectedPilotCount := 1
				pilotCount := len(pilots.Items)
				if pilotCount != expectedPilotCount {
					t.Log(pretty.Sprint(pilots))
					t.Errorf("Unexpected pilot count: %d != %d", expectedPilotCount, pilotCount)
				}
			},
		},
	}
	for title, test := range tests {
		t.Run(
			title,
			func(t *testing.T) {
				fixture := &framework.StateFixture{
					T:                t,
					KubeObjects:      test.kubeObjects,
					NavigatorObjects: test.navObjects,
				}
				fixture.Start()
				defer fixture.Stop()
				state := fixture.State()
				c := pilot.NewControl(
					state.NavigatorClientset,
					state.PilotLister,
					state.PodLister,
					state.StatefulSetLister,
					state.Recorder,
				)
				err := c.Sync(test.cluster)
				if err != nil {
					if !test.expectErr {
						t.Errorf("Unexpected error: %s", err)
					} else {
						t.Logf("The following error was expected: %s", err)
					}
				} else {
					if test.expectErr {
						t.Error("Missing error")
					}
				}
				if test.assertions != nil {
					test.assertions(t, state)
				}
			},
		)
	}
}
