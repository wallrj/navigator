/*
Copyright 2017 Jetstack Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This file was automatically generated by informer-gen

package internalversion

import (
	navigator "github.com/jetstack-experimental/navigator/pkg/apis/navigator"
	internalclientset "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/internalclientset"
	internalinterfaces "github.com/jetstack-experimental/navigator/pkg/client/informers_generated/internalversion/internalinterfaces"
	internalversion "github.com/jetstack-experimental/navigator/pkg/client/listers_generated/navigator/internalversion"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
	time "time"
)

// ElasticsearchClusterInformer provides access to a shared informer and lister for
// ElasticsearchClusters.
type ElasticsearchClusterInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() internalversion.ElasticsearchClusterLister
}

type elasticsearchClusterInformer struct {
	factory internalinterfaces.SharedInformerFactory
}

// NewElasticsearchClusterInformer constructs a new informer for ElasticsearchCluster type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewElasticsearchClusterInformer(client internalclientset.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				return client.Navigator().ElasticsearchClusters(namespace).List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				return client.Navigator().ElasticsearchClusters(namespace).Watch(options)
			},
		},
		&navigator.ElasticsearchCluster{},
		resyncPeriod,
		indexers,
	)
}

func defaultElasticsearchClusterInformer(client internalclientset.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewElasticsearchClusterInformer(client, v1.NamespaceAll, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
}

func (f *elasticsearchClusterInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&navigator.ElasticsearchCluster{}, defaultElasticsearchClusterInformer)
}

func (f *elasticsearchClusterInformer) Lister() internalversion.ElasticsearchClusterLister {
	return internalversion.NewElasticsearchClusterLister(f.Informer().GetIndexer())
}
