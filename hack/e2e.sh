#!/bin/bash
set -eux


: ${TEST_PREFIX:=""}

: ${NAVIGATOR_IMAGE_REPOSITORY:="jetstackexperimental"}
: ${NAVIGATOR_IMAGE_TAG:="build"}
: ${NAVIGATOR_IMAGE_PULLPOLICY:="Never"}

export \
    NAVIGATOR_IMAGE_REPOSITORY \
    NAVIGATOR_IMAGE_TAG \
    NAVIGATOR_IMAGE_PULLPOLICY

NAVIGATOR_NAMESPACE="navigator"
RELEASE_NAME="nav-e2e"

ROOT_DIR="$(git rev-parse --show-toplevel)"
SCRIPT_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd)"
CONFIG_DIR=$(mktemp -d -t navigator-e2e.XXXXXXXXX)
mkdir -p $CONFIG_DIR
CERT_DIR="$CONFIG_DIR/certs"
mkdir -p $CERT_DIR
TEST_DIR="$CONFIG_DIR/tmp"
mkdir -p $TEST_DIR

source "${SCRIPT_DIR}/libe2e.sh"

# Mandatory environment variables
: ${CHART_VALUES:?}
: ${CHART_VALUES_CASSANDRA:?}

helm delete --purge "${RELEASE_NAME}" || true

function debug_navigator_start() {
    kubectl api-versions
    kubectl get pods --all-namespaces
    kubectl describe deploy
    kubectl describe pod
}

function helm_install() {
    helm delete --purge "${RELEASE_NAME}" || true
    echo "Installing navigator..."
    if helm --debug install --wait --name "${RELEASE_NAME}" contrib/charts/navigator \
         --values ${CHART_VALUES}
    then
        return 0
    fi
    return 1
}

# Retry helm install to work around intermittent API server availability.
# See https://github.com/jetstack/navigator/issues/118
if ! retry helm_install; then
    debug_navigator_start
    echo "ERROR: Failed to install Navigator"
    exit 1
fi

# Wait for navigator API to be ready
function navigator_ready() {
    local replica_count_controller=$(
        kubectl get deployment ${RELEASE_NAME}-navigator-controller \
                --output 'jsonpath={.status.readyReplicas}' || true)
    if [[ "${replica_count_controller}" -eq 0 ]]; then
        return 1
    fi
    local replica_count_apiserver=$(
        kubectl get deployment ${RELEASE_NAME}-navigator-apiserver \
                --output 'jsonpath={.status.readyReplicas}' || true)
    if [[ "${replica_count_apiserver}" -eq 0 ]]; then
        return 1
    fi
    if ! kubectl api-versions | grep 'navigator.jetstack.io'; then
        return 1
    fi
    # Even after the API appears in api-versions, it takes a short time for API
    # server to recognise navigator API types.
    if ! kubectl get esc; then
        return 1
    fi
    if ! kube_event_exists "kube-system" \
         "navigator-controller:Endpoints:Normal:LeaderElection"
    then
        return 1
    fi
    return 0
}

echo "Waiting for Navigator to be ready..."
if ! retry navigator_ready; then
    debug_navigator_start
    echo "ERROR: Timeout waiting for Navigator API"
    exit 1
fi

FAILURE_COUNT=0
TEST_ID="$(date +%s)-${RANDOM}"

function fail_test() {
    FAILURE_COUNT=$(($FAILURE_COUNT+1))
    echo "TEST FAILURE: $1"
}

function test_elasticsearchcluster() {
    local namespace="${1}"
    echo "Testing ElasticsearchCluster"
    kubectl create namespace "${namespace}"
    if ! kubectl get esc; then
        fail_test "Failed to use shortname to get ElasticsearchClusters"
    fi
    # Create and delete an ElasticSearchCluster
    if ! kubectl create \
            --namespace "${namespace}" \
            --filename \
            <(envsubst \
                  '$NAVIGATOR_IMAGE_REPOSITORY:$NAVIGATOR_IMAGE_TAG:$NAVIGATOR_IMAGE_PULLPOLICY' \
                  < "${SCRIPT_DIR}/testdata/es-cluster-test.template.yaml")
    then
        fail_test "Failed to create elasticsearchcluster"
    fi
    if ! kubectl get \
            --namespace "${namespace}" \
            ElasticSearchClusters; then
        fail_test "Failed to get elasticsearchclusters"
    fi
    if ! retry kubectl get \
         --namespace "${namespace}" \
         service es-test; then
        fail_test "Navigator controller failed to create elasticsearchcluster service"
    fi
    if ! retry kube_event_exists "${namespace}" \
         "navigator-controller:ElasticsearchCluster:Normal:SuccessSync"
    then
        fail_test "Navigator controller failed to create SuccessSync event"
    fi
    if ! retry TIMEOUT=300 stdout_gt 0 kubectl \
         --namespace "${namespace}" \
         get pilot \
         "es-test-mixed-0" \
         "-o=go-template={{.status.elasticsearch.documents}}"
    then
        fail_test "Elasticsearch pilot did not update the document count"
    fi
}

if [[ "test_elasticsearchcluster" = "${TEST_PREFIX}"* ]]; then
    ES_TEST_NS="test-elasticsearchcluster-${TEST_ID}"
    test_elasticsearchcluster "${ES_TEST_NS}"
    if [ "${FAILURE_COUNT}" -gt "0" ]; then
        fail_and_exit "${ES_TEST_NS}"
    fi
    kube_delete_namespace_and_wait "${ES_TEST_NS}"
fi

function cql_connect() {
    local namespace="${1}"
    local host="${2}"
    local port="${3}"
    # Attempt to negotiate a CQL connection.
    # No queries are performed.
    # stdin=false (the default) ensures that cqlsh does not go into interactive
    # mode.
    # XXX: This uses the standard Cassandra Docker image rather than the
    # gcr.io/google-samples/cassandra image used in the Cassandra chart, becasue
    # cqlsh is missing some dependencies in that image.
    kubectl \
        run \
        "cql-responding-${RANDOM}" \
        --namespace="${namespace}" \
        --command=true \
        --image=cassandra:latest \
        --restart=Never \
        --rm \
        --stdin=false \
        --attach=true \
        -- \
        /usr/bin/cqlsh --debug "${host}" "${port}"
}

function test_cassandracluster() {
    echo "Testing CassandraCluster"
    local namespace="${1}"
    local CHART_NAME="cassandra-${TEST_ID}"

    kubectl create namespace "${namespace}"

    if ! kubectl get \
         --namespace "${namespace}" \
         CassandraClusters; then
        fail_test "Failed to get cassandraclusters"
    fi

    helm install \
         --debug \
         --wait \
         --name "${CHART_NAME}" \
         --namespace "${namespace}" \
         contrib/charts/cassandra \
         --values "${CHART_VALUES_CASSANDRA}" \
         --set replicaCount=1

    # Wait 5 minutes for cassandra to start and listen for CQL queries.
    if ! retry TIMEOUT=300 cql_connect \
         "${namespace}" \
         "cass-${CHART_NAME}-cassandra-cql" \
         9042; then
        fail_test "Navigator controller failed to create cassandracluster service"
    fi

    # TODO Fail test if there are unexpected cassandra errors.
    kubectl log \
            --namespace "${namespace}" \
            "statefulset/cass-${CHART_NAME}-cassandra-ringnodes"

    # Change the CQL port
    helm --debug upgrade \
         "${CHART_NAME}" \
         contrib/charts/cassandra \
         --values "${CHART_VALUES_CASSANDRA}" \
         --set replicaCount=1 \
         --set cqlPort=9043

    # Wait 60s for cassandra CQL port to change
    if ! retry TIMEOUT=60 cql_connect \
         "${namespace}" \
         "cass-${CHART_NAME}-cassandra-cql" \
         9043; then
        fail_test "Navigator controller failed to update cassandracluster service"
    fi

    # Increment the replica count
    helm --debug upgrade \
         "${CHART_NAME}" \
         contrib/charts/cassandra \
         --values "${CHART_VALUES_CASSANDRA}" \
         --set cqlPort=9043 \
         --set replicaCount=2

    if ! retry TIMEOUT=300 stdout_equals 2 kubectl \
         --namespace "${namespace}" \
         get statefulsets \
         "cass-${CHART_NAME}-cassandra-ringnodes" \
         "-o=go-template={{.status.readyReplicas}}"
    then
        fail_test "Second cassandra node did not become ready"
    fi

    simulate_unresponsive_cassandra_process \
        "${namespace}" \
        "cass-${CHART_NAME}-cassandra-ringnodes-0" \
        "cassandra"

    if ! retry cql_connect \
         "${namespace}" \
         "cass-${CHART_NAME}-cassandra-cql" \
         9043; then
        fail_test "Cassandra readiness probe failed to bypass dead node"
    fi
}

if [[ "test_cassandracluster" = "${TEST_PREFIX}"* ]]; then
    CASS_TEST_NS="test-cassandra-${TEST_ID}"
    test_cassandracluster "${CASS_TEST_NS}"
    if [ "${FAILURE_COUNT}" -gt "0" ]; then
        fail_and_exit "${CASS_TEST_NS}"
    fi
    kube_delete_namespace_and_wait "${CASS_TEST_NS}"
fi
