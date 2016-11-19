FROM arangodb/arangodb
MAINTAINER Max Neunhoeffer <max@arangodb.com>

ADD ./entry.sh /entry.sh
ADD ./FindTasks/FindTasks /FindTasks

ENTRYPOINT /entry.sh
CMD []
