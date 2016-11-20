package main

import (
	"bytes"
	//"encoding/json"
	"errors"
	"flag"
  "fmt"
	"io/ioutil"
	"net/http"
	"time"
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
  ]%s
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

var residencySkeleton string = `,
  "residency": {
    "relaunchEscalationTimeoutSeconds": 3600,
    "taskLostBehavior": "WAIT_FOREVER"
  }
`

func makeAgencyJSON() (bufAll bytes.Buffer) {
	bufAll = bytes.Buffer{}
	bufVol := bytes.Buffer{}
	fmt.Fprintf(&bufVol, volumeSkeleton, agentDiskLimit)
	fmt.Fprintf(&bufAll, allSkeleton, clusterName, "agency", agentCPULimit,
	            agentMemLimit, agentDiskLimit, agentNumber,
							string(bufVol.Bytes()), agentNumber, constraintSkeleton,
						  residencySkeleton);
	return
}

func makeCoordinatorJSON() (bufAll bytes.Buffer) {
	bufAll = bytes.Buffer{}
	fmt.Fprintf(&bufAll, allSkeleton, clusterName, "coordinators",
	            coordinatorCPULimit, coordinatorMemLimit, coordinatorDiskLimit,
							coordinatorNumber, "", agentNumber, "", "");
	return
}

func makeDBServerJSON() (bufAll bytes.Buffer) {
	bufAll = bytes.Buffer{}
	bufVol := bytes.Buffer{}
	fmt.Fprintf(&bufVol, volumeSkeleton, dbserverDiskLimit)
	fmt.Fprintf(&bufAll, allSkeleton, clusterName, "dbservers", dbserverCPULimit,
	            dbserverMemLimit, dbserverDiskLimit, dbserverNumber,
							string(bufVol.Bytes()), agentNumber, constraintSkeleton,
						  residencySkeleton);
	return
}

func checkDeployment(instancetype string, maker func() bytes.Buffer) error {
	r, e := http.Get(marathonURL + "/v2/apps/" + clusterName)
	if e != nil || r == nil {
		fmt.Println("Error contacting Marathon for type", instancetype, ":", e)
		return e
	}
	r.Body.Close()
	if r.StatusCode == http.StatusOK {
		fmt.Println("Found Marathon deployment for type", instancetype, ", good.")
		return nil
	}
	json := maker()
	fmt.Println("Trying to POST to Marathon:", string(json.Bytes()))
	r, e = http.Post(marathonURL + "/v2/apps", "application/json", &json)
	if e != nil || r == nil {
		fmt.Println("Error POSTing to Marathon for type", instancetype, ":", e)
		return e
	}
	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if r.StatusCode != http.StatusCreated {
		fmt.Println("Error response from Marathon for type", instancetype, ":",
		            r.StatusCode, string(body))
		return errors.New("Error response from Marathon:" + string(body))
	}
	return nil
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
	count := 0
	for {
		time.Sleep(10000000000)
		count++
		fmt.Printf("Living %d\n", count)
	}
}
