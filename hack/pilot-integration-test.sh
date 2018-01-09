#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

RELEASE_NAME="nav-e2e"
NAMESPACE="bar"

: ${CHART_VALUES:?}

ROOT_DIR="$(git rev-parse --show-toplevel)"
SCRIPT_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd)"
source "${SCRIPT_DIR}/libe2e.sh"

touch cleanup.tmp
function cleanup() {
    source cleanup.tmp || true
    > cleanup.tmp
}
cleanup
trap cleanup EXIT

go build -o "${SCRIPT_DIR}/pilot-cassandra" "./cmd/pilot-cassandra"

docker build --tag pilot-cassandra:test --file "${SCRIPT_DIR}/Dockerfile.pilot-integration-test" "${SCRIPT_DIR}"

install_navigator_and_wait "${RELEASE_NAME}"

kube_delete_namespace_and_wait "${NAMESPACE}" || true
kubectl create namespace "${NAMESPACE}"
echo kubectl delete --now namespace "${NAMESPACE}" >> cleanup.tmp

kubectl proxy -v4 --port 8001 >logs.proxy 2>&1 &
proxy_pid=$!
echo kill "${proxy_pid}" >> cleanup.tmp
echo wait "${proxy_pid}" >> cleanup.tmp

retry TIMEOUT=1 curl http://localhost:8001

docker run \
       --name pilot-under-test \
       --detach \
       --net host \
       --volume "${PWD}:/etc/pilot:ro" \
       pilot-cassandra:test \
       --alsologtostderr \
       --logtostderr \
       --v 6 \
       --master http://localhost:8001 \
       --pilot-name foo \
       --pilot-namespace bar \
       --config-dir /etc/pilot
echo docker rm --force pilot-under-test >> cleanup.tmp


# Check the health endpoints
retry curl -v http://localhost:12000
retry curl -v http://localhost:12001

kubectl create --namespace bar --filename "${SCRIPT_DIR}/pilot.yaml"
echo kubectl delete --filename "${ROOT_DIR}/hack/pilot.yaml" --namespace "${NAMESPACE}" --now >> cleanup.tmp

retry stdout_equals "ProcessStarted" kubectl get pilot foo --namespace bar --output 'jsonpath={.status.conditions[0].reason}'
