package main

import (
	"bytes"
	"flag"
  "fmt"
)

var clusterName string
var marathonURL string
var agentCPULimit float64
var agentMemLimit uint
var agentDiskLimit uint
var agentNumber uint
var dbserverCPULimit float64
var dbserverMemLimit uint
var dbserverDiskLimit uint
var dbserverNumber uint
var coordinatorCPULimit float64
var coordinatorMemLimit uint
var coordinatorDiskLimit uint
var coordinatorNumber uint

var allSkeleton string = `{
  "id": "%s%s",
  "cpus": %.3f,
  "mem": %d,
  "disk": %d,
  "instances": %d,
  "container": {
    "docker": {
      "image": "neunhoef/arangodb-marathon",
      "network": "BRIDGE",
      "forcePullImage": false,
      "portMappings": [
        {
          "containerPort": 8529,
          "hostPort": 0,
          "protocol": "tcp"
        }
      ]
    },
    "type": "DOCKER"%s
  },
  "env": {
    "AGENCY_SIZE": "%d"
  },
  %s"labels": {},
  "healthChecks": [
    { "path": "/_api/version", "portIndex": 0, "protocol": "HTTP",
      "gracePeriodSeconds": 300, "intervalSeconds": 30,
      "timeoutSeconds": 5, "maxConsecutiveFailures": 0,
      "ignoreHttp1xx": false }
  ],
  "residency": {
    "relaunchEscalationTimeoutSeconds": 3600,
    "taskLostBehavior": "WAIT_FOREVER"
  }
}
`

var volumeSkeleton string = `,
    "volumes": [
      {
        "hostPath": "data",
        "containerPath": "/var/lib/arangodb3",
        "mode": "RW"
      },
      {
        "containerPath": "data",
        "persistent": {
          "size": %d
        },
        "mode": "RW"
      }
    ]
`

var constraintSkeleton string = `"constraints": [["hostname", "UNIQUE"]],
  `

func makeAgencyJSON() []byte {
	bufAll := bytes.Buffer{}
	bufVol := bytes.Buffer{}
	fmt.Fprintf(&bufVol, volumeSkeleton, agentDiskLimit)
	fmt.Fprintf(&bufAll, allSkeleton, clusterName, "agency", agentCPULimit,
	            agentMemLimit, agentDiskLimit, agentNumber,
							string(bufVol.Bytes()), agentNumber, constraintSkeleton);
	return bufAll.Bytes()
}

func makeCoordinatorJSON() []byte {
	bufAll := bytes.Buffer{}
	fmt.Fprintf(&bufAll, allSkeleton, clusterName, "coordinators",
	            coordinatorCPULimit, coordinatorMemLimit, coordinatorDiskLimit,
							coordinatorNumber, "", agentNumber, "");
	return bufAll.Bytes()
}

func makeDBServerJSON() []byte {
	bufAll := bytes.Buffer{}
	bufVol := bytes.Buffer{}
	fmt.Fprintf(&bufVol, volumeSkeleton, dbserverDiskLimit)
	fmt.Fprintf(&bufAll, allSkeleton, clusterName, "dbservers", dbserverCPULimit,
	            dbserverMemLimit, dbserverDiskLimit, dbserverNumber,
							string(bufVol.Bytes()), agentNumber, constraintSkeleton);
	return bufAll.Bytes()
}

func checkDeployment(instancetype string, maker func() []byte) {
	// GET $MARATHON/v2/apps/${NAME}${instancetype}
	// if there: return
	// make JSON using makerfunc
	// POST $MARATHON/v2/apps
	// if worked: return
	// log error
	fmt.Println(string(maker()))
	return
}

func checkDeployments() {
	checkDeployment("agency", makeAgencyJSON)
	checkDeployment("dbservers", makeDBServerJSON)
	checkDeployment("coordinators", makeCoordinatorJSON)
}

func main() {
	flag.StringVar(&clusterName, "name", "arangodb", "name of ArangoDB cluster")
	flag.StringVar(&marathonURL, "marathon", "http://marathon.mesos:8080",
	               "URL to Marathon")
  flag.Float64Var(&agentCPULimit, "agentCPULimit", 1.0,
	                "CPU resource limit for agents")
	flag.UintVar(&agentMemLimit, "agentMemLimit", 2048,
	             "Memory resource limit for agents")
	flag.UintVar(&agentDiskLimit, "agentDiskLimit", 2048,
	             "Memory resource limit for agents")
  flag.UintVar(&agentNumber, "agentNumber", 3, "Number of agents in agency")
  flag.Float64Var(&dbserverCPULimit, "dbserverCPULimit", 1.0,
	                "CPU resource limit for dbservers")
	flag.UintVar(&dbserverMemLimit, "dbserverMemLimit", 2048,
	             "Memory resource limit for DBservers")
	flag.UintVar(&dbserverDiskLimit, "dbserverDiskLimit", 2048,
	             "Memory resource limit for dbservers")
  flag.UintVar(&dbserverNumber, "dbserverNumber", 2, "Number of DBservers")
  flag.Float64Var(&coordinatorCPULimit, "coordinatorCPULimit", 1.0,
	                "CPU resource limit for coordinators")
	flag.UintVar(&coordinatorMemLimit, "coordinatorMemLimit", 2048,
	             "Memory resource limit for coordinators")
	flag.UintVar(&coordinatorDiskLimit, "coordinatorDiskLimit", 2048,
	             "Memory resource limit for coordinators")
  flag.UintVar(&coordinatorNumber, "coordinatorNumber", 2,
	             "Number of coordinators")
	flag.Parse()

	checkDeployments()
}
