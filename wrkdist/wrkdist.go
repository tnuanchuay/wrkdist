package wrkdist

import (
	"regexp"
	"fmt"
	"strings"
	"strconv"
	"os/exec"
	"bufio"
	"runtime"
)

type WrkResult struct {
	JobID		uint
	IsError		bool
	Url		string
	Duration	float64
	Thread		int
	Connection	int
	Latency_Avg	float64
	Latency_Stdev	float64
	Latency_Max	float64
	ReqPerSec_Avg	float64
	ReqPerSec_Stdev	float64
	ReqPerSec_Max	float64
	Requests	int
	RequestPerSec	float64
	TransferPerSec		float64
	TotalTransfer		float64
	SocketErrors_Connection	int
	SocketErrors_Read	int
	SocketErrors_Write	int
	SocketErrors_Timeout	int
	Non2xx3xx	int
	TestcaseID		uint
}

func Run(url, c, d string) WrkResult{
	var command *exec.Cmd

	command = exec.Command("wrk", fmt.Sprintf("-t%d", runtime.NumCPU()) , fmt.Sprintf("-c%s", c), fmt.Sprintf("-d%s", d) ,url)

	fmt.Println(command.Args)
	cmdReader, _ := command.StdoutPipe()
	scanner := bufio.NewScanner(cmdReader)
	defer cmdReader.Close()
	var out string
	go func() {
		for scanner.Scan() {
			out = fmt.Sprintf("%s\n%s", out, scanner.Text())
			if strings.Contains(out, "Transfer"){
				break;
			}
		}
	}()
	command.Start()
	command.Wait()

	fmt.Println(out)

	wrk := WrkResult{}
	wrk.SetData(url, out)
	return wrk
}

func (wrkResult *WrkResult) SetData(url, out string){
	wrkResult.Url = url
	wrkResult.SetDuration(out)
	wrkResult.SetThread(out)
	wrkResult.SetConnection(out)
	wrkResult.SetRequestPerSec(out)
	wrkResult.SetRequests(out)
	wrkResult.SetTransferPerSec(out)
	wrkResult.SetLatency(out)
	wrkResult.SetReqPerSec(out)
	wrkResult.SetTotalTransfer(out)
	wrkResult.SetSocketErrors(out)
}

func (t *WrkResult) SetSocketErrors(s string){
	regexpErr1 := regexp.MustCompile("Socket errors: connect [0-9]*, read [0-9]*, write [0-9]*, timeout [0-9]*")
	result := regexpErr1.FindAllStringSubmatch(s, -1)
	if len(result) == 1{
		textError1 := result[0][0]
		textError1 = strings.Replace(textError1, ",", "", -1)
		splitedTextError1 := strings.Fields(textError1)
		t.SocketErrors_Connection, _ = strconv.Atoi(splitedTextError1[3])
		t.SocketErrors_Read, _ = strconv.Atoi(splitedTextError1[5])
		t.SocketErrors_Write, _ = strconv.Atoi(splitedTextError1[7])
		t.SocketErrors_Timeout, _ = strconv.Atoi(splitedTextError1[9])
	}

	regexpErr2 := regexp.MustCompile("Non-2xx or 3xx responses: [0-9]*")
	result = regexpErr2.FindAllStringSubmatch(s, -1)
	if len(result) == 1{
		textError2 := result[0][0]
		splitedTextError2 := strings.Fields(textError2)[4]
		t.Non2xx3xx, _ = strconv.Atoi(splitedTextError2)
	}
}

func (t *WrkResult) SetTotalTransfer(s string){
	regexpTotalTransfer := regexp.MustCompile(", [0-9A-Za-z.]* read")
	result := regexpTotalTransfer.FindAllStringSubmatch(s, -1)
	if len(result) != 1{
		t.IsError = true
		fmt.Println("TotalTransfer Error")
	}else{
		textTotalTransfer := result[0][0]
		splitedTextTotalTransfer := strings.Fields(textTotalTransfer)
		t.TotalTransfer,_ = SIToFloat(splitedTextTotalTransfer[1])
		//fmt.Println("t.TotalTransfer", t.TotalTransfer)
	}
}

func (t *WrkResult) SetReqPerSec(s string){
	reqexpReqPerSec := regexp.MustCompile("Req/Sec[ ]*[0-9A-Za-z.]*[ ]*[0-9A-Za-z.]*[ ]*[0-9A-Za-z.]*")
	result := reqexpReqPerSec.FindAllStringSubmatch(s, -1)
	if len(result) != 1{
		t.IsError = true
		fmt.Println("ReqPerSec Error")
	}else{
		textReqPerSec := result[0][0]
		sqlitedTextReqPerSec := strings.Fields(textReqPerSec)
		t.ReqPerSec_Avg, _ = SIToFloat(sqlitedTextReqPerSec[1])
		t.ReqPerSec_Stdev, _ = SIToFloat(sqlitedTextReqPerSec[2])
		t.ReqPerSec_Max, _ = SIToFloat(sqlitedTextReqPerSec[3])
	}
}

func (t *WrkResult) SetLatency(s string){
	regexpLatency := regexp.MustCompile("Latency[ ]*[0-9A-Za-z.]*[ ]*[0-9A-Za-z.]*[ ]*[0-9A-Za-z.]*")
	result := regexpLatency.FindAllStringSubmatch(s, -1)
	if len(result) != 1{
		fmt.Println("Latency Error")
		t.IsError = true
	}else{
		textLatency := result[0][0]
		splitedTextLatency := strings.Fields(textLatency)
		t.Latency_Avg, _ = TimeToFloat(splitedTextLatency[1])
		t.Latency_Stdev, _ = TimeToFloat(splitedTextLatency[2])
		t.Latency_Max, _ = TimeToFloat(splitedTextLatency[3])
	}
}

func (t *WrkResult) SetTransferPerSec(s string){
	regexpTps := regexp.MustCompile("Transfer/sec:[ ]*[0-9.]*[KMG]B")
	result := regexpTps.FindAllStringSubmatch(s, -1)
	if len(result) != 1{
		fmt.Println("TransferPerSec Error")
		t.IsError = true
	}else{
		textTps := result[0][0]
		splitedTextTps := strings.Fields(textTps)
		t.TransferPerSec, _ = SIToFloat(splitedTextTps[len(splitedTextTps) - 1])
		//fmt.Println("t.TransferPerSec", t.TransferPerSec)
	}
}

func (t *WrkResult) SetRequestPerSec(s string){
	regexpRps := regexp.MustCompile("Requests/sec:[ ]*[0-9.]*")
	result := regexpRps.FindAllStringSubmatch(s, -1)
	if len(result) != 1{
		t.IsError = true
		fmt.Println("RequestPerSec Error")
	}else{
		textRps := result[0][0]
		splitedTextRps := strings.Fields(textRps)
		t.RequestPerSec, _ = strconv.ParseFloat(splitedTextRps[len(splitedTextRps) - 1], 64)
		//fmt.Println("t.RequestPerSec", t.RequestPerSec)
	}
}

func (t *WrkResult) SetRequests(s string){
	regexpRps := regexp.MustCompile("[0-9]* requests")
	result := regexpRps.FindAllStringSubmatch(s, -1)

	if len(result) != 1{
		t.IsError = true
		fmt.Println("Request Error")
	}else{
		textReq := result[0][0]
		splitedTestReq := strings.Fields(textReq)[0]
		t.Requests, _ = strconv.Atoi(splitedTestReq)
		//fmt.Println("t.Requests", t.Requests)
	}
}

func (t *WrkResult) SetDuration(s string){
	regexpDuration := regexp.MustCompile("requests in [0-9A-Za-z.]*,")
	result := regexpDuration.FindAllStringSubmatch(s, -1)

	if len(result) != 1{
		fmt.Println("Duration Error")
		t.IsError = true
	}else{
		textTime := result[0][0]
		textTime = strings.Replace(textTime, ",", "", -1)
		splitedTextTime := strings.Fields(textTime)[2]
		t.Duration, _ = TimeToFloat(splitedTextTime)
		//fmt.Println("t.duration", t.Duration)
	}
}

func (t *WrkResult) SetThread(s string){
	regexpThread := regexp.MustCompile("[0-9]* threads")
	result := regexpThread.FindAllStringSubmatch(string(s), -1)

	if len(result) != 1{
		fmt.Println("Thread Error")
		t.IsError = true
	}else{
		textThread := result[0][0]
		splitedTextThread := strings.Fields(textThread)[0]
		threadNum, _ := SIToFloat(splitedTextThread)
		t.Thread = int(threadNum)
		//fmt.Println("t.Thread", t.Thread)
	}
}

func (t *WrkResult) SetConnection(s string){
	regexpConnection := regexp.MustCompile("[0-9]* connections")
	result := regexpConnection.FindAllStringSubmatch(s, -1)

	if len(result) != 1{
		t.IsError = true
		fmt.Println("Connection Error")
	}else{
		textConnection := result[0][0]
		splitedTextConnection := strings.Fields(textConnection)[0]
		threadNum, _ := SIToFloat(splitedTextConnection)
		t.Connection = int(threadNum)
		//fmt.Println("t.Connection", t.Connection)
	}
}
