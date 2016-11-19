ArangoDB Docker image to be run under Mesos via Marathon
========================================================

The idea is that this image can turn into an arbitrary type of instance
(agent, dbserver or coordinator), only depending on an environment variable.
All necessary command line options for arangod are generated within the
image, solely based on the following environment variables:

  - `MARATHON_APP_ID`: must be set and end in `agency` or `coordinators` 
    or `dbservers`

Optional:

  - `AGENCY_SIZE`: must be an odd number, defaults to 3
