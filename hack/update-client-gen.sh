#!/bin/bash

# The only argument this script should ever be called with is '--verify-only'

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE}")/..
BINDIR=${REPO_ROOT}/bin
HACKDIR=${HACKDIR:-hack}

# Generate the internal clientset (pkg/client/clientset_generated/internalclientset)
${BINDIR}/client-gen "$@" \
	      --input-base "github.com/jetstack-experimental/navigator/pkg/apis/" \
	      --input "navigator/" \
	      --clientset-path "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/" \
	      --clientset-name internalclientset \
	      --go-header-file "${HACKDIR}/boilerplate.go.txt"
# Generate the versioned clientset (pkg/client/clientset_generated/clientset)
${BINDIR}/client-gen "$@" \
		  --input-base "github.com/jetstack-experimental/navigator/pkg/apis/" \
		  --input "navigator/v1alpha1" \
	      --clientset-path "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/" \
	      --clientset-name "clientset" \
	      --go-header-file "${HACKDIR}/boilerplate.go.txt"
# generate lister
${BINDIR}/lister-gen "$@" \
		  --input-dirs="github.com/jetstack-experimental/navigator/pkg/apis/navigator" \
	      --input-dirs="github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1" \
	      --output-package "github.com/jetstack-experimental/navigator/pkg/client/listers_generated" \
	      --go-header-file "${HACKDIR}/boilerplate.go.txt"
# generate informer
${BINDIR}/informer-gen "$@" \
	      --go-header-file "${HACKDIR}/boilerplate.go.txt" \
	      --input-dirs "github.com/jetstack-experimental/navigator/pkg/apis/navigator" \
	      --input-dirs "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1" \
	      --internal-clientset-package "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/internalclientset" \
	      --versioned-clientset-package "github.com/jetstack-experimental/navigator/pkg/client/clientset_generated/clientset" \
	      --listers-package "github.com/jetstack-experimental/navigator/pkg/client/listers_generated" \
	      --output-package "github.com/jetstack-experimental/navigator/pkg/client/informers_generated"
