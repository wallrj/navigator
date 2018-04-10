package validation_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jetstack/navigator/pkg/apis/navigator"
	"github.com/jetstack/navigator/pkg/apis/navigator/validation"
	"github.com/jetstack/navigator/pkg/cassandra/version"
)

var (
	validCassCluster = &navigator.CassandraCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "bar",
		},
		Spec: navigator.CassandraClusterSpec{
			Version: *version.New("3.11.2"),
			Image:   &validImageSpec,
			NavigatorClusterConfig: validNavigatorClusterConfig,
			NodePools: []navigator.CassandraClusterNodePool{
				navigator.CassandraClusterNodePool{
					Datacenter:  "datacenter-1",
					Rack:        "rack-1",
					Persistence: validNodePoolPersistenceConfig,
				},
			},
		},
	}
)

func TestValidateCassandraCluster(t *testing.T) {
	type testT struct {
		cluster       *navigator.CassandraCluster
		errorExpected bool
	}

	setVersion := func(
		c *navigator.CassandraCluster,
		v *version.Version,
	) *navigator.CassandraCluster {
		c = c.DeepCopy()
		c.Spec.Version = *v
		return c
	}

	tests := map[string]testT{
		"valid cluster": {
			cluster: validCassCluster,
		},
		"version too low": {
			cluster:       setVersion(validCassCluster, version.New("2.0.0")),
			errorExpected: true,
		},
		"version too high": {
			cluster:       setVersion(validCassCluster, version.New("4.0.0")),
			errorExpected: true,
		},
	}

	setNavigatorClusterConfig := func(
		c *navigator.CassandraCluster,
		ncc navigator.NavigatorClusterConfig,
	) *navigator.CassandraCluster {
		c = c.DeepCopy()
		c.Spec.NavigatorClusterConfig = ncc
		return c
	}

	for title, ncc := range navigatorClusterConfigErrorCases {
		tests[title] = testT{
			cluster:       setNavigatorClusterConfig(validCassCluster, ncc),
			errorExpected: true,
		}
	}

	setImage := func(
		c *navigator.CassandraCluster,
		image *navigator.ImageSpec,
	) *navigator.CassandraCluster {
		c = c.DeepCopy()
		c.Spec.Image = image
		return c
	}

	for title, image := range imageErrorCases {
		tests[title] = testT{
			cluster:       setImage(validCassCluster, &image),
			errorExpected: true,
		}
	}

	for title, tc := range tests {
		t.Run(
			title,
			func(t *testing.T) {
				errs := validation.ValidateCassandraCluster(tc.cluster)
				if tc.errorExpected && len(errs) == 0 {
					t.Errorf("expected error but got none")
				}
				if !tc.errorExpected && len(errs) != 0 {
					t.Errorf("unexpected errors: %s", errs)
				}
				for _, e := range errs {
					t.Logf("error string is: %s", e)
				}
			},
		)
	}
}

func TestValidateCassandraClusterUpdate(t *testing.T) {
	type testT struct {
		old           *navigator.CassandraCluster
		new           *navigator.CassandraCluster
		errorExpected bool
	}

	setPersistence := func(
		c *navigator.CassandraCluster,
		p navigator.PersistenceConfig,
	) *navigator.CassandraCluster {
		c = c.DeepCopy()
		c.Spec.NodePools[0].Persistence = p
		return c
	}

	setRack := func(
		c *navigator.CassandraCluster,
		rack string,
	) *navigator.CassandraCluster {
		c = c.DeepCopy()
		c.Spec.NodePools[0].Rack = rack
		return c
	}

	setDatacenter := func(
		c *navigator.CassandraCluster,
		rack string,
	) *navigator.CassandraCluster {
		c = c.DeepCopy()
		c.Spec.NodePools[0].Datacenter = rack
		return c
	}

	setVersion := func(
		c *navigator.CassandraCluster,
		v *version.Version,
	) *navigator.CassandraCluster {
		c = c.DeepCopy()
		c.Spec.Version = *v
		return c
	}

	tests := map[string]testT{
		"unchanged cluster": {
			old: validCassCluster,
			new: validCassCluster,
		},
		"changed rack": {
			old:           validCassCluster,
			new:           setRack(validCassCluster, "toot"),
			errorExpected: true,
		},
		"changed datacenter": {
			old:           validCassCluster,
			new:           setDatacenter(validCassCluster, "doot"),
			errorExpected: true,
		},
		"enable persistence config": {
			old: setPersistence(validCassCluster, navigator.PersistenceConfig{Enabled: false}),
			new: validCassCluster,
		},
		"downgrade not allowed": {
			old:           setVersion(validCassCluster, validCassCluster.Spec.Version.BumpMinor()),
			new:           validCassCluster,
			errorExpected: true,
		},
		"upgrade not allowed": {
			old:           validCassCluster,
			new:           setVersion(validCassCluster, validCassCluster.Spec.Version.BumpMinor()),
			errorExpected: true,
		},
	}

	for title, persistence := range persistenceErrorCases {
		tests[title] = testT{
			old:           validCassCluster,
			new:           setPersistence(validCassCluster, persistence),
			errorExpected: true,
		}
	}

	for title, tc := range tests {
		t.Run(
			title,
			func(t *testing.T) {
				errs := validation.ValidateCassandraClusterUpdate(tc.old, tc.new)
				if tc.errorExpected && len(errs) == 0 {
					t.Errorf("expected error but got none")
				}
				if !tc.errorExpected && len(errs) != 0 {
					t.Errorf("unexpected errors: %s", errs)
				}
				for _, e := range errs {
					t.Logf("error string is: %s", e)
				}
			},
		)
	}
}
