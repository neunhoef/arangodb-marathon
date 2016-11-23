#/bin/bash

set -e

env

sleep 1 

if [ -z "$MARATHON_APP_ID" ] ; then
    echo "This Docker image is supposed to be started via Marathon."
    sleep 30
    exit 1
fi
# Some defaults:
AGENCY_SIZE=${AGENCY_SIZE:-3}
MARATHON=${MARATHON:-"http://marathon.mesos:8080"}

case "$MARATHON_APP_ID" in
    */servers/dbservers) 
       CLUSTER_NAME=${MARATHON_APP_ID%/servers/dbservers}
       TYPE=dbserver
       OPTION="--cluster.agency-endpoint"
       ;;
    */servers/coordinators)
        CLUSTER_NAME=${MARATHON_APP_ID%/servers/coordinators}
        TYPE=coordinator
        OPTION="--cluster.agency-endpoint"
        ;;
    *agency/agents)
        CLUSTER_NAME=${MARATHON_APP_ID%/agency/agents}
        TYPE=agent
        OPTION="--agency.endpoint"
        ;;
    *)
        echo "The MARATHON_APP_ID is invalid: $MARATHON_APP_ID"
        sleep 30
        exit 2
        ;;
esac

CLUSTER_NAME=${CLUSTER_NAME}
echo CLUSTER_NAME=$CLUSTER_NAME

AGENTS=$(/FindTasks -marathon $MARATHON -oneline -option=${OPTION} -prefix=tcp:// -minimum $AGENCY_SIZE ${CLUSTER_NAME}/agency/agents)

echo Agents found: $AGENTS

mkdir -p /var/lib/arangodb3
mkdir -p /var/lib/arangodb3-apps
chown -R arangodb /var/lib/arangodb3
chown -R arangodb /var/lib/arangodb3-apps

case "$TYPE" in
    dbserver)
        echo Starting DBserver...
        exec /usr/sbin/arangod --cluster.my-address tcp://${HOST}:${PORT0} \
                               --server.endpoint tcp://0.0.0.0:8529 \
                               --cluster.my-local-info ${HOST}:${PORT0} \
                               --cluster.my-role PRIMARY \
                               --server.authentication false \
                               $AGENTS "$@"
        ;;
    coordinator)
        echo Starting Coordinator...
        exec /usr/sbin/arangod --cluster.my-address tcp://${HOST}:${PORT0} \
                               --server.endpoint tcp://0.0.0.0:8529 \
                               --cluster.my-local-info ${HOST}:${PORT0} \
                               --cluster.my-role COORDINATOR \
                               --server.authentication false \
                               $AGENTS "$@"
        ;;
    agent)
        echo Starting Agent...
        exec /usr/sbin/arangod --agency.activate true \
                               --agency.my-address tcp://${HOST}:${PORT0} \
                               --agency.size ${AGENCY_SIZE} \
                               --agency.supervision true \
                               --server.endpoint tcp://0.0.0.0:8529 \
                               --server.statistics false \
                               --server.authentication false \
                               $AGENTS "$@"
        ;;
esac

