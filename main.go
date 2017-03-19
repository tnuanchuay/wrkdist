package main

import (
	"os"
	"flag"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"fmt"
	"errors"
	"github.com/parnurzeal/gorequest"
	"net/http"
	"time"
	"github.com/tspn/wrkdist/wrkdist"
	"strconv"
	"crypto/sha1"
)

const(
	WRKDISTJSON		=		"./wrkdist.json"
	WRKRESULTJSON		=		"./wrkresult.json"
)

const (
	MODE_INIT		=		"init"
	MODE_ADD		=		"add"
	MODE_DEL		=		"del"
	MODE_RUN		=		"run"
	MODE_LIST		=		"list"
	MODE_WORKER		=		"worker"

	FLAG_CONNECTION		=		"c"
	FLAG_DURATION		=		"d"

	MODE_TASKLIST		=		"task-list"
	MODE_TASKSUM		=		"task-sum"
)

type Node struct {
	Ip		net.IP
	Status		string
	Message		string
}

type Task struct{
	ID		string
	Start		time.Time
	Summary		wrkdist.WrkResult
}

type TaskResultFile struct {
	Task	[]wrkdist.WrkResult
}

type Setting struct{
	Node	[]Node
	Task	[]Task
}

type StatusResponse struct {
	Status		string
}

type TaskResponse struct {
	Status		bool
	WrkResult	wrkdist.WrkResult
}

const (
	ECONFIGNOTFOUND 			=		"Please initial setting file by use init"
	EPARSEIP				=		"Cannot parse IP Address"
	EPARSEJSON				=		"Cannot parse setting json file."
	ENOHOSTAVAILABLE			=		"No Node Available."
	ENEEDWRKPARAM				=		"Need --c and --d params."
	ETASKNOTFOUND				=		"Task not Found."
)

//wrkdist init
//wrkdist add
//wrkdist del
//wrkdist run
//wrkdist list

const (
	WORKERIDLE				=		"IDLE"
	WORKERRUNNING				=		"RUNNING"
	WORKERCOOLDOWN				=		"COOLDOWN"
	WORKERDEAD				=		"DEAD"
)

var workerState = WORKERIDLE
var task map[string]wrkdist.WrkResult = make(map[string]wrkdist.WrkResult)

type RequestToRun struct {
	TaskID		string
	Url		string
	Thread		string
	Connection	string
	Duration	string
}

func main(){
	isInitMode := flag.Bool(MODE_INIT, false, "Initial state.")
	isAddMode := flag.Bool(MODE_ADD, false, "Add node by ipv4 into the node pool.")
	isDelMode := flag.Bool(MODE_DEL, false, "Del node from node pool.")
	isRunMode := flag.Bool(MODE_RUN, false, "Run wrk all of the node for result.")
	isListMode := flag.Bool(MODE_LIST, false, "List all node format ipv4 in node pool.")
	isWorkerMode := flag.Bool(MODE_WORKER, false, "Run as worker mode.")
	isTaskModeList := flag.Bool(MODE_TASKLIST, false, "List all Task.")
	isTaskModeSum := flag.Bool(MODE_TASKSUM, false, "Read Summary Task Result.")

	runModeConnection := flag.String(FLAG_CONNECTION, "", "Number of Connection.")
	runModeDuration := flag.String(FLAG_DURATION, "", "Test Duration.")

	flag.Parse()

	switch {
	case *isInitMode != false:
		initMode()
	case *isAddMode != false:
		addMode()
	case *isDelMode != false:
		delMode()
	case *isRunMode != false:
		if runModeDuration == nil || runModeConnection == nil{
			log.Fatal(ENEEDWRKPARAM)
		}
		runMode(*runModeConnection, *runModeDuration, getLastArg())
	case *isWorkerMode != false:
		workerMode()
	case *isListMode != false:
		listMode()
	case *isTaskModeList != false:
		taskListFunc()
	case *isTaskModeSum != false:
		taskSumFunc()

	default:
		fmt.Println(os.Args, "command not found.")
	}
}

func taskSumFunc() {
	fileExist := isConfigFileExist()
	if !fileExist {
		log.Fatal(ECONFIGNOTFOUND)
	}

	config := readSetting()
	var taskResponse []TaskResponse

	for _, node := range config.Node{
		task := getTaskFromNode(node.Ip, getLastArg())
		if task.Status {
			taskResponse = append(taskResponse, task)
		}
	}

	sumResult := wrkdist.WrkResult{}

	if len(taskResponse) > 0{
		for _, task := range taskResponse{
			sumResult.Latency_Avg += task.WrkResult.Latency_Avg
			sumResult.Latency_Max += task.WrkResult.Latency_Max
			sumResult.Latency_Stdev += task.WrkResult.Latency_Stdev
			sumResult.RequestPerSec += task.WrkResult.RequestPerSec
			sumResult.Requests += task.WrkResult.Requests
			sumResult.Non2xx3xx += task.WrkResult.Non2xx3xx
			sumResult.SocketErrors_Connection += task.WrkResult.SocketErrors_Connection
			sumResult.SocketErrors_Read += task.WrkResult.SocketErrors_Read
			sumResult.SocketErrors_Timeout += task.WrkResult.SocketErrors_Timeout
			sumResult.SocketErrors_Write += task.WrkResult.SocketErrors_Write
			sumResult.TotalTransfer += task.WrkResult.TotalTransfer
			sumResult.TransferPerSec += task.WrkResult.TransferPerSec
			sumResult.Connection += task.WrkResult.Connection
			sumResult.Duration += task.WrkResult.Duration
		}

		sumResult.Latency_Avg /= float64(len(taskResponse))
		sumResult.Latency_Max /= float64(len(taskResponse))
		sumResult.Latency_Stdev /= float64(len(taskResponse))
		sumResult.Duration /= float64(len(taskResponse))

		fmt.Println("Connection\t\t\t\t\t=", sumResult.Connection)
		fmt.Println("Duration\t\t\t\t\t=", sumResult.Duration)
		fmt.Println("Requests/Sec\t\t\t\t\t=", sumResult.RequestPerSec)
		fmt.Println("Tranfers/Sec\t\t\t\t\t=", sumResult.TransferPerSec)
		fmt.Println("Requests\t\t\t\t\t=", sumResult.Requests)
		fmt.Println("Total Transfer\t\t\t\t\t=", sumResult.TotalTransfer)
		fmt.Println("Average Latency\t\t\t\t\t=", sumResult.Latency_Avg)
		fmt.Println("Max Latency\t\t\t\t\t=", sumResult.Latency_Max)
		fmt.Println("Standard Deviation Latency\t\t\t=", sumResult.Latency_Stdev)
		fmt.Println("Non 2xx 3xx\t\t\t\t\t=", sumResult.Non2xx3xx)
		fmt.Println("Connection Socket Error\t\t\t\t=", sumResult.SocketErrors_Connection)
		fmt.Println("Write Socket Error\t\t\t\t=", sumResult.SocketErrors_Write)
		fmt.Println("Read Socket Error\t\t\t\t=", sumResult.SocketErrors_Read)
		fmt.Println("Timeout Socket Error\t\t\t\t=", sumResult.SocketErrors_Timeout)

	}else{
		log.Fatal(ETASKNOTFOUND)
	}
}

func getTaskFromNode(ips net.IP, id string) TaskResponse{
	taskResponse := TaskResponse{}

	req := gorequest.New()
	url := urlResult(ips, id)
	res, body, err := req.Get(url).End()
	//log.Fatal(url)
	if err == nil{
		if res.StatusCode == 200{
			err := json.Unmarshal([]byte(body), &taskResponse)
			if err != nil {
				log.Fatal(err)
			}
		}else{
			log.Println("Err response from", ips, "status code", res.StatusCode)
		}
	}
	return taskResponse
}

func taskListFunc() {
	config := readSetting()
	fmt.Println("ID", "\t\t\t\t", "Start Time")
	for _, item := range config.Task {
		fmt.Println(item.ID, "\t\t\t",item.Start)
	}
}

func runMode(c, d, url string) {
	fileExist := isConfigFileExist()
	if !fileExist {
		log.Fatal(ECONFIGNOTFOUND)
	}

	config := readSetting()
	id := generateTaskId()

	pingAllInList(&config)

	var idleNodes []Node
	for _, node := range config.Node{
		if node.Status == WORKERIDLE{
			idleNodes = append(idleNodes, node)
		}
	}

	if !(len(idleNodes) > 0){
		log.Fatal(ENOHOSTAVAILABLE)
	}

	cFloat, err := wrkdist.SIToFloat(c)
	if err != nil{
		log.Fatal(err)
	}

	eachNodeConnection := int(cFloat) / len(idleNodes)
	reqToRun := RequestToRun{TaskID:id, Connection:strconv.Itoa(eachNodeConnection), Duration:d, Url:url}

	for _, node := range idleNodes{
		requestToNode(urlRun(node.Ip), reqToRun)
	}

	config.Task = append(config.Task, Task{ID:id, Start:time.Now()})
	saveConfigFile(config)
}

func workerMode() {
	fileExist := isResultFileExist()
	if !fileExist {
		createNewResultFile()
	}

	resultFile := readResult()
	for _, item := range resultFile.Task{
		task[item.TaskID] = item
	}

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request){
		if r.Method == "GET" {
			responseData, err := json.Marshal(StatusResponse{Status:workerState})
			if err != nil {
				log.Fatal(err)
			}

			fmt.Fprint(w, string(responseData))
		}
	})

	http.HandleFunc("/wrk", func(w http.ResponseWriter, r *http.Request){
		if r.Method == "GET"{
			id := r.FormValue("id")
			if id != ""{
				wrkResult := task[id]
				if wrkResult.TaskID == "" {
					res, _ := json.Marshal(TaskResponse{Status:false})
					fmt.Fprint(w, string(res))
				}

				byteJsonWrkResult, err := json.Marshal(TaskResponse{Status:true, WrkResult:wrkResult})
				if err != nil {
					log.Println(err)
				}

				fmt.Fprint(w, string(byteJsonWrkResult))
			}

		}else if r.Method == "POST"{
			if workerState == WORKERIDLE {
				decoder := json.NewDecoder(r.Body)
				requestForRun := RequestToRun{}
				err := decoder.Decode(&requestForRun)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Println(requestForRun)
				workerState = WORKERRUNNING

				go func() {
					task[requestForRun.TaskID] = wrkdist.Run(requestForRun.TaskID, requestForRun.Url, requestForRun.Connection, requestForRun.Duration)
					resultFile.Task = append(resultFile.Task, task[requestForRun.TaskID])
					saveResultFile(resultFile)
					workerState = WORKERCOOLDOWN
					time.Sleep(60 * time.Second)
					workerState = WORKERIDLE
				}()
			}

			responseData, err := json.Marshal(StatusResponse{Status:workerState})
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(responseData))
			fmt.Fprint(w, string(responseData))
		}
	})

	http.ListenAndServe(":12321", nil)
}

func requestToNode(url string, reqToRun RequestToRun){
	gorequest.New().Post(url).Send(reqToRun).End()
}

func generateTaskId() string {
	sha := sha1.New()
	sha.Write([]byte(time.Now().Format(time.RFC3339)))
	idByte := sha.Sum(nil)

	return fmt.Sprintf("%x", idByte[:5])
}

func listMode() {
	fileExist := isConfigFileExist()
	if !fileExist {
		log.Fatal(ECONFIGNOTFOUND)
	}

	config := readSetting()

	pingAllInList(&config)
	listAllNodeStatus(config)

	saveConfigFile(config)
}

func pingAllInList(config *Setting){
	nodes := (*config).Node
	for i := 0 ; i < len(nodes) ; i++{
		nodes[i].Status = ping(nodes[i].Ip)
	}
}

func listAllNodeStatus(config Setting) {
	fmt.Println("#", "\t\t\t", "IP ADDRESS", "\t\t\t", "STATUS")
	for i, item := range config.Node {
		fmt.Println(i, "\t\t\t", item.Ip.String(), "\t\t\t",item.Status)
	}
}

func delMode() {
	fileExist := isConfigFileExist()
	if !fileExist{
		log.Fatal(ECONFIGNOTFOUND)
	}

	config := readSetting()
	delNode(&config, getLastArg())
	saveConfigFile(config)
}

func addMode() {
	fileExist := isConfigFileExist()
	if !fileExist{
		log.Fatal(ECONFIGNOTFOUND)
	}

	config := readSetting()
	ip := getLastArg()
	addNode(&config, ip)
	saveConfigFile(config)
}

func initMode() {
	fileExist := isConfigFileExist()
	if !fileExist {
		createNewSettingFile()
	}else{
		fmt.Println("Setting file already exist.")
	}
}

func delNode(setting *Setting, arg string) {
	ip := net.ParseIP(arg)
	ip = ip.To4()
	if ip == nil {
		log.Fatal(EPARSEIP)
	}
	index, err := nodePoolIndexOf(*setting, arg)
	if err != nil {
		log.Fatal("Not found this ip in node pool")
	}

	(*setting).Node = append((*setting).Node[:index], (*setting).Node[index+1:]...)
}

func getLastArg() string{
	return os.Args[len(os.Args)-1]
}

func saveConfigFile(setting Setting) {
	jsonSetting, err := json.Marshal(setting)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(WRKDISTJSON, jsonSetting, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func addNode(setting *Setting, ipString string) {
	ip := net.ParseIP(ipString)
	ip = ip.To4()
	if ip == nil {
		log.Fatal(EPARSEIP)
	}

	if isExistNode(setting, ip){
		log.Fatal("Node exist in the node pool.")
	}

	status := ping(ip)
	if status == WORKERDEAD{
		fmt.Printf("Cannot connect to %s but it will be in the pool.\n", ipString)
	}

	(*setting).Node = append((*setting).Node, Node{Ip:ip, Status:status})
}

func ping(ips net.IP) string{
	status := WORKERDEAD

	req := gorequest.New()
	url := urlStat(ips)
	res, body, err := req.Get(url).End()
	if err == nil{
		if res.StatusCode == 200{
			resStatus := StatusResponse{}
			err := json.Unmarshal([]byte(body), &resStatus)
			status = resStatus.Status
			if err != nil{
				log.Fatal(EPARSEJSON)
			}
		}else{
			log.Println("Err response from", ips, "status code", res.StatusCode)
		}
	}
	return status
}

func urlStat(ip net.IP) string{
	return fmt.Sprintf("http://%s:12321/status", ip.String())
}

func urlRun(ip net.IP) string{
	return fmt.Sprintf("http://%s:12321/wrk", ip.String())
}

func urlResult(ip net.IP, id string) string{
	return fmt.Sprintf("http://%s:12321/wrk?id=%s", ip.String(), id)
}

func isExistNode(setting *Setting, ip net.IP) bool{
	for _, item := range setting.Node{
		if item.Ip.Equal(ip) {
			return true
		}
	}

	return false
}

func nodePoolIndexOf(setting Setting, ipArg string)(int, error){
	ip := net.ParseIP(ipArg)
	ip.To4()

	if ip == nil {
		log.Fatal(EPARSEIP)
	}

	for i, item := range setting.Node{
		if item.Ip.Equal(ip) {
			return  i, nil
		}
	}

	return 0, errors.New("not found")
}

func readSetting() Setting {
	config := Setting{}

	file, err := ioutil.ReadFile(WRKDISTJSON)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatal(err)
	}

	return config
}

func createNewSettingFile() {
	setting := Setting{}
	settingJson, err := json.Marshal(setting)
	if err != nil {
		log.Fatal(EPARSEJSON)
	}

	ioutil.WriteFile(WRKDISTJSON, settingJson, 0644)
}

func isConfigFileExist() bool {
	fileExist := true
	if _, err := os.Stat(WRKDISTJSON); os.IsNotExist(err){
		fileExist = false
	}
	return fileExist
}

func isResultFileExist() bool {
	fileExist := true
	if _, err := os.Stat(WRKRESULTJSON); os.IsNotExist(err){
		fileExist = false
	}
	return fileExist
}

func createNewResultFile() {
	resultFile := TaskResultFile{}
	resultFileJson, err := json.Marshal(resultFile)
	if err != nil {
		log.Fatal(EPARSEJSON)
	}

	ioutil.WriteFile(WRKRESULTJSON, resultFileJson, 0644)
}

func readResult() TaskResultFile {
	tasks := TaskResultFile{}

	file, err := ioutil.ReadFile(WRKRESULTJSON)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(file, &tasks)
	if err != nil {
		log.Fatal(err)
	}

	return tasks
}

func saveResultFile(resultFile TaskResultFile) {
	jsonresultFile, err := json.Marshal(resultFile)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(WRKRESULTJSON, jsonresultFile, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

