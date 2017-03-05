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
)

type Server struct {
	Ip		net.IP
}

type Setting struct{
	Node	[]Server
}

const (
	ECONFIGNOTFOUND 			=		"Please initial setting file by use init"
	EPARSEIP				=		"Cannot parse IP Address"
)

func main(){
	isInitMode := flag.Bool(MODE_INIT, false, "Initial state.")
	isAddMode := flag.Bool(MODE_ADD, false, "Add node by ipv4 into the node pool.")
	isDelMode := flag.Bool(MODE_DEL, false, "Del node from node pool.")
	isRunMode := flag.Bool(MODE_RUN, false, "Run wrk all of the node for result")
	isListMode := flag.Bool(MODE_LIST, false, "List all node format ipv4 in node pool")

	flag.Parse()

	switch {
	case *isInitMode != false:
		fileExist := isConfigFileExist()
		if !fileExist {
			createNewSettingFile()
		}else{
			fmt.Println("Setting file already exist.")
		}
	case *isAddMode != false:
		fileExist := isConfigFileExist()
		if !fileExist{
			log.Fatal(ECONFIGNOTFOUND)
		}

		config := readSetting()
		addNode(&config, getLastArg())
		saveConfigFile(config)
	case *isDelMode != false:
		fileExist := isConfigFileExist()
		if !fileExist{
			log.Fatal(ECONFIGNOTFOUND)
		}

		config := readSetting()
		delNode(&config, getLastArg())
		saveConfigFile(config)
	case *isRunMode != false:
	case *isListMode != false:
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
	
	(*setting).Node = append((*setting).Node, Server{Ip:ip})
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
		log.Fatal("Cannot parse setting json file.")
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