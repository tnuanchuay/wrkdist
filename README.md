# WRKDIST

 Distribution Wrk for high connection Load Testing.
 
 ### Installation
 ```
 go get github.com/tspn/wrkdist
 ```
 
 ### Usage
 - initial project
 ```
 wrkdist --init
 ```
 - add host/server to node pool
 ```
 wrkdist --add 127.0.0.1
 ```
 - remove host/server from node pool
 ```
 wrkdist --del 127.0.0.1
 ```
 - list node pool
 ```
 wrkdist --list
 ```
 - run
 ```
 wrkdist --run --c=50k --d=100s http://127.0.0.1
 ```
 - list recently task/job/run
 ```
 wrkdist --task-list
 ```
 - get result
 ```
 wrkdist --task-sum <#id>
 ```
