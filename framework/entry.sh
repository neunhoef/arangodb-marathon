#!/bin/bash
env
MARATHON_APP_ID=${MARATHON_APP_ID:-arangodb}
exec /arangodb-framework -name ${MARATHON_APP_ID} "$@"
