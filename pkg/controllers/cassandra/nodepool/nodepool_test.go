package nodepool_test

import (
	"testing"

	casstesting "github.com/jetstack/navigator/pkg/controllers/cassandra/testing"
)

func TestNodePoolControlSync(t *testing.T) {
	t.Run(
		"create a statefulset",
		func(t *testing.T) {
			f := casstesting.NewFixture(t)
			f.Run()
			f.AssertStatefulSetsLength(1)
		},
	)
	// t.Run(
	//	"ignore existing statefulset",
	//	func(t *testing.T) {
	//		f := casstesting.NewFixture(t)
	//		f.AddObjectK(
	//			nodepool.StatefulSetForCluster(
	//				f.Cluster,
	//				&f.Cluster.Spec.NodePools[0],
	//			),
	//		)
	//		f.Run()
	//		f.AssertStatefulSetsLength(1)
	//	},
	// )
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
