package nodepool_test

import (
	"testing"

	"github.com/jetstack/navigator/pkg/apis/navigator/v1alpha1"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/nodepool"
	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
)

func TestNodePoolControlSync(t *testing.T) {
	t.Run(
		"create a statefulset",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			status := f.Run()
			f.AssertStatefulSetsLength(1)
			if len(status.NodePools) != 0 {
				t.Errorf("Expected no nodepool status. Found: %#v", status.NodePools)
			}
		},
	)
	t.Run(
		"add NodePoolStatus if a matching StatefulSet exists",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.AddObjectK(
				nodepool.StatefulSetForCluster(
					f.Cluster,
					&f.Cluster.Spec.NodePools[0],
				),
			)
			status := f.Run()
			f.AssertStatefulSetsLength(1)
			if len(status.NodePools) != 1 {
				t.Errorf("Expected one nodepool status. Found: %#v", status.NodePools)
			}
		},
	)
	t.Run(
		"update NodePoolStatus.ReadyReplicas to match StatefulSet.ReadyReplicas",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			np := &f.Cluster.Spec.NodePools[0]
			ss := nodepool.StatefulSetForCluster(
				f.Cluster,
				np,
			)
			ss.Status.ReadyReplicas = np.Replicas
			f.AddObjectK(ss)
			status := f.Run()
			f.AssertStatefulSetsLength(1)
			if np.Replicas != status.NodePools[np.Name].ReadyReplicas {
				t.Errorf(
					"Unexpected NodePoolStatus.ReadyReplicas: %d != %d",
					np.Replicas,
					status.NodePools[np.Name].ReadyReplicas,
				)
			}
		},
	)
	t.Run(
		"remove NodePoolStatus if no matching StatefulSet exists",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			np := f.Cluster.Spec.NodePools[0]
			f.Cluster.Status.NodePools = map[string]v1alpha1.CassandraClusterNodePoolStatus{
				np.Name: {},
			}
			status := f.Run()
			if _, found := status.NodePools[np.Name]; found {
				t.Error("Orphan NodePoolStatus was not deleted:", status)
			}
		},
	)

	t.Run(
		"remove NodePoolStatus if no matching NodePool exists",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.Cluster.Status.NodePools = map[string]v1alpha1.CassandraClusterNodePoolStatus{
				"orphan-status-1234": {},
			}
			status := f.Run()
			f.AssertStatefulSetsLength(1)
			if _, found := status.NodePools["orphan-status-1234"]; found {
				t.Error("Orphan NodePoolStatus was not deleted:", status)
			}
		},
	)
	// t.Run(
	//	"update statefulset",
	//	func(t *testing.T) {
	//		f := casstesting.NewFixture(t)
	//		unsyncedSet := nodepool.StatefulSetForCluster(
	//			f.Cluster,
	//			&f.Cluster.Spec.NodePools[0],
	//		)
	//		unsyncedSet.SetLabels(map[string]string{})
	//		f.AddObjectK(unsyncedSet)
	//		f.Run()
	//		f.AssertStatefulSetsLength(1)
	//		sets := f.StatefulSets()
	//		set := sets.Items[0]
	//		labels := set.GetLabels()
	//		if len(labels) == 0 {
	//			t.Log(set)
	//			t.Error("StatefulSet was not updated")
	//		}
	//	},
	// )
	// t.Run(
	//	"error on update foreign statefulset",
	//	func(t *testing.T) {
	//		f := casstesting.NewFixture(t)
	//		foreignUnsyncedSet := nodepool.StatefulSetForCluster(
	//			f.Cluster,
	//			&f.Cluster.Spec.NodePools[0],
	//		)
	//		foreignUnsyncedSet.SetLabels(map[string]string{})
	//		foreignUnsyncedSet.OwnerReferences = nil
	//		f.AddObjectK(foreignUnsyncedSet)
	//		f.RunExpectError()
	//	},
	// )
	// t.Run(
	//	"delete statefulset without nodepool",
	//	func(t *testing.T) {
	//		f := casstesting.NewFixture(t)
	//		f.AddObjectK(
	//			nodepool.StatefulSetForCluster(
	//				f.Cluster,
	//				&f.Cluster.Spec.NodePools[0],
	//			),
	//		)
	//		f.Cluster.Spec.NodePools = []v1alpha1.CassandraClusterNodePool{}
	//		f.Run()
	//		f.AssertStatefulSetsLength(0)
	//	},
	// )
	// t.Run(
	//	"do not delete foreign owned stateful sets",
	//	func(t *testing.T) {
	//		f := casstesting.NewFixture(t)
	//		foreignStatefulSet := nodepool.StatefulSetForCluster(
	//			f.Cluster,
	//			&f.Cluster.Spec.NodePools[0],
	//		)
	//		foreignStatefulSet.OwnerReferences = nil

	//		f.AddObjectK(foreignStatefulSet)
	//		f.Cluster.Spec.NodePools = []v1alpha1.CassandraClusterNodePool{}
	//		f.RunExpectError()
	//	},
	// )
}
