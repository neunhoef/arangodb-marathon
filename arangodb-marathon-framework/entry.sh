#!/bin/bash
env
MARATHON_APP_ID=${MARATHON_APP_ID:-/arangodb-framework}
CLUSTER_NAME=${MARATHON_APP_ID%-framework}
exec /arangodb-marathon-framework -name ${CLUSTER_NAME} "$@"
