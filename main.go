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
)

const(
	WRKDISTJSON		=		"./wrkdist.json"
)

const (
	MODE_INIT		=		"init"
	MODE_ADD		=		"add"
	MODE_DEL		=		"del"
	MODE_RUN		=		"run"
	MODE_LIST		=		"list"
	MODE_WORKER		=		"worker"
)

const (
	OK			=		"OK"
	DEAD			=		"DEAD"
	RUNNING			=		"RUNNING"
	COOLDOWN		=		"COOLDOWN"
)

type Server struct {
	Ip		net.IP
	Status		string
	Message		string
}

type Setting struct{
	Node	[]Server
}

type StatusResponse struct {
	Status		string
}

const (
	ECONFIGNOTFOUND 			=		"Please initial setting file by use init"
	EPARSEIP				=		"Cannot parse IP Address"
	EPARSEJSON				=		"Cannot parse setting json file."
)

//wrkdist init
//wrkdist add
//wrkdist del
//wrkdist run
//wrkdist list

func main(){
	isInitMode := flag.Bool(MODE_INIT, false, "Initial state.")
	isAddMode := flag.Bool(MODE_ADD, false, "Add node by ipv4 into the node pool.")
	isDelMode := flag.Bool(MODE_DEL, false, "Del node from node pool.")
	isRunMode := flag.Bool(MODE_RUN, false, "Run wrk all of the node for result.")
	isListMode := flag.Bool(MODE_LIST, false, "List all node format ipv4 in node pool.")
	isWorkerMode := flag.Bool(MODE_WORKER, false, "Run as worker mode.")
	flag.Parse()

	switch {
	case *isInitMode != false:
		initMode()
	case *isAddMode != false:
		addMode()
	case *isDelMode != false:
		delMode()
	case *isRunMode != false:
	case *isWorkerMode != false:
	case *isListMode != false:
		listMode()
	default:
		fmt.Println(os.Args, "command not found.")
	}
}

func listMode() {
	fileExist := isConfigFileExist()
	if !fileExist {
		log.Fatal(ECONFIGNOTFOUND)
	}

	config := readSetting()

	pingAll(config)
	listAllNodeStatus(config)

	saveConfigFile(config)
}
func pingAll(config Setting) {
	for _, item := range config.Node {
		item.Status = ping(item.Ip)
	}
}

func listAllNodeStatus(config Setting) {
	fmt.Println("#", "\t\t\t", "ip", "\t\t\t", "\tstatus")
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
	if status == DEAD{
		fmt.Printf("Cannot connect to %s but it will be in the pool.\n", ipString)
	}

	(*setting).Node = append((*setting).Node, Server{Ip:ip, Status:status})
}

func ping(ips net.IP) string{
	status := DEAD

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
	return fmt.Sprintf("https://%s:12321/stat", ip.String())
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