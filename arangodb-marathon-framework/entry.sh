#!/bin/bash
env
MARATHON_APP_ID=${MARATHON_APP_ID:-arangodb}
exec /arangodb-marathon-framework -name ${MARATHON_APP_ID} "$@"
