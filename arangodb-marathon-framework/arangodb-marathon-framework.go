package main

import (
	"bytes"
	//"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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
  "id": "%s",
  "groups": [
	  {
      "id": "%s/agency",
      "apps": [
%s
      ]
    },
    {
      "id": "%s/servers",
			"dependencies": [ "%s/agency" ],
      "apps": [
%s,
%s
      ]
    }
  ]
}
`

var appSkeleton string = `{
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
          "protocol": "tcp"%s
        }
      ]
    },
    "type": "DOCKER"%s
  },
  "env": {
    "AGENCY_SIZE": "%d"
  },
  %s"labels": {
  },
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

var minuteManSkeleton string = `,
          "labels": {"VIP_0": "%s:8529"},
					"name": "%s"`

func makeAgencyJSON() (bufAll bytes.Buffer) {
	bufAll = bytes.Buffer{}
	bufVol := bytes.Buffer{}
	fmt.Fprintf(&bufVol, volumeSkeleton, agentDiskLimit)
	fmt.Fprintf(&bufAll, appSkeleton, clusterName, "/agency/agents",
	  agentCPULimit, agentMemLimit, agentDiskLimit, agentNumber, "",
		string(bufVol.Bytes()), agentNumber, constraintSkeleton,
		residencySkeleton)
	return
}

func makeCoordinatorJSON() (bufAll bytes.Buffer) {
	bufAll = bytes.Buffer{}
	bufMin := bytes.Buffer{}
	strippedClusterName := clusterName[1:]
	fmt.Fprintf(&bufMin, minuteManSkeleton, strippedClusterName,
	            strippedClusterName)
	fmt.Fprintf(&bufAll, appSkeleton, clusterName, "/servers/coordinators",
		coordinatorCPULimit, coordinatorMemLimit, coordinatorDiskLimit,
		coordinatorNumber, string(bufMin.Bytes()), "", agentNumber,
		"", "")
	return
}

func makeDBServerJSON() (bufAll bytes.Buffer) {
	bufAll = bytes.Buffer{}
	bufVol := bytes.Buffer{}
	fmt.Fprintf(&bufVol, volumeSkeleton, dbserverDiskLimit)
	fmt.Fprintf(&bufAll, appSkeleton, clusterName, "/servers/dbservers",
	  dbserverCPULimit, dbserverMemLimit, dbserverDiskLimit, dbserverNumber, "",
		string(bufVol.Bytes()), agentNumber, constraintSkeleton,
		residencySkeleton)
	return
}

func checkDeployment(json bytes.Buffer) error {
	r, e := http.Get(marathonURL + "/v2/groups" + clusterName)
	if e != nil || r == nil {
		fmt.Println("Error contacting Marathon:", e)
		return e
	}
	r.Body.Close()
	if r.StatusCode == http.StatusOK {
		fmt.Println("Found Marathon deployment for name", clusterName, ", good.")
		return nil
	}
	fmt.Println("Trying to POST to Marathon:", string(json.Bytes()))
	r, e = http.Post(marathonURL+"/v2/groups", "application/json", &json)
	if e != nil || r == nil {
		fmt.Println("Error POSTing to Marathon:", e)
		return e
	}
	body, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if r.StatusCode != http.StatusCreated {
		fmt.Println("Error response from Marathon:", r.StatusCode, string(body))
		return errors.New("Error response from Marathon:" + string(body))
	}
	return nil
}

func checkDeployments() {
	agencyJSON := makeAgencyJSON()
  dbserverJSON := makeDBServerJSON()
	coordinatorJSON := makeCoordinatorJSON()
  bufAll := bytes.Buffer{}
	fmt.Fprintf(&bufAll, allSkeleton, clusterName, clusterName,
	  string(agencyJSON.Bytes()), clusterName, clusterName,
		string(dbserverJSON.Bytes()), string(coordinatorJSON.Bytes()))
	checkDeployment(bufAll)
}

func serveStatus(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"ok": true}`))
}

func serveShutdown(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got shutdown request, trying to talk to Marathon...")

	tr := &http.Transport{
	  DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 5,
	}
	client := &http.Client{Transport: tr}
	req, _ := http.NewRequest("DELETE", marathonURL + "/v2/groups" + clusterName,
                            nil)
	rr, ee := client.Do(req)
	if ee != nil || rr == nil {
		fmt.Println("Error contacting Marathon:", ee)
	  w.Write([]byte(`{"ok": false}`))
		return
	}
	rr.Body.Close()
	if rr.StatusCode == http.StatusOK {
		fmt.Println("Deleted Marathon deployment for name", clusterName, ", good.")
	  w.Write([]byte(`{"ok": true}`))
		return
	}
	w.Write([]byte(`{"ok": false}`))
}

func serveHttp() {
	http.HandleFunc("/v2/status", serveStatus)
	http.HandleFunc("/v2/shutdown", serveShutdown)
	port, found := os.LookupEnv("PORT0")
	if !found {
		port = "8000"
	}
	fmt.Println("Serving HTTP/REST API on port", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
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

	go serveHttp()

	checkDeployments()
	count := 0
	for {
		time.Sleep(10000000000)
		count++
		fmt.Printf("Living %d\n", count)
	}
}
