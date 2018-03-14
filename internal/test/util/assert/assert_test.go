package assert_test

import (
	"testing"

	"github.com/jetstack/navigator/internal/test/util/assert"
	"github.com/jetstack/navigator/internal/test/util/generate"
	"github.com/jetstack/navigator/pkg/controllers/cassandra/util"
	"github.com/kr/pretty"
)

func TestObjectAttributes(t *testing.T) {
	o1 := generate.StatefulSet(
		generate.StatefulSetConfig{
			Name:      "set1",
			Namespace: "ns1",
		},
	)
	o2 := generate.StatefulSet(
		generate.StatefulSetConfig{
			Name:      "set1",
			Namespace: "ns1",
			Replicas:  util.Int32Ptr(3),
		},
	)
	var nilReplicas *int32
	assert.ObjectMatches(
		t,
		o1,
		map[string]interface{}{
			"Name":          "set1",
			"Namespace":     "ns1",
			"Spec.Replicas": nilReplicas,
		},
	)
	res := pretty.Diff(o1, o2)
	for key, v := range res {
		t.Log("key", key)
		t.Log("v", v)
	}

}
