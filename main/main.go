package main

import "github.com/tspn/wrk-load-testing-module/wrk"

func main(){
	
	wrk.Run("http://127.0.0.1", "1", "1", "1s")
}
