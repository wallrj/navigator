#!/bin/bash

# dind-cluster can't pull from private registries

set -o errexit
set -o nounset
set -o pipefail

function docker_login_from_node() {
    local node=$1
    local token=$2
    docker exec "${node}" docker login --username oauth2accesstoken --password "${token}" https://gcr.io
}

function docker_pull_images_from_node() {
    local node=$1
    for component in apiserver controller pilot-elasticsearch pilot-cassandra;
    do
        docker exec "${node}" docker pull "gcr.io/jetstack-sandbox/navigator-${component}:build"
    done
}

function main() {
    local nodes=$(kubectl get nodes -o 'jsonpath={.items[*].metadata.name}')
    local token=$(gcloud auth application-default print-access-token)
    local images=""
    for n in $nodes
    do
        docker_login_from_node "${n}" "${token}"
        docker_pull_images_from_node "${n}"
    done
}

main
