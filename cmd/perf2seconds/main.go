// perf2Seconds reduces a perf .csv file to a by-seconds perf .csv file
package main

//import (
//	"bufio"
//	"fmt"
//	"log"
//	"os"
//	"os/exec"
//)
//
func main() {
	//	// parse the comamnd-line
	//	//get an input or -
	//	// open it as a file
	//	// sort it and pipe to an fd
	//	// this admittedly looks wierd, use --debug to see what it does
	//	// sort -k 2.1,2.2nb -k 2.5,2.6nb -k2.8,2.12nb --temporary-directory=/var/tmp |\
	//	// call perf2Seconds(fd)
	//
	//	//// docker build current directory
	//	cmdName := "sort"
	//	cmdArgs := []string{"-k", "2.1,2.2nb", "-k", "2.5,2.6nb", "-k", "2.8,2.12nb",
	//		"--temporary-directory=/var/tmp"}
	//
	//	cmd := exec.Command(cmdName, cmdArgs...)
	//	cmdReader, err := cmd.StdoutPipe()
	//	if err != nil {
	//		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
	//		os.Exit(1)
	//	}
	//	scanner := bufio.NewScanner(cmdReader)
	//	go func() {
	//		for scanner.Scan() {
	//			fmt.Printf("docker build out | %s\n", scanner.Text())
	//		}
	//	}()
	//
	//	err = cmd.Start()
	//	if err != nil {
	//		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
	//		os.Exit(1)
	//	}
	//
	//	err = cmd.Wait()
	//	if err != nil {
	//		fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
	//		os.Exit(1)
	//	}
	//	file, err := os.Open("/path/to/file.txt")
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	defer file.Close()
	//
	//	scanner := bufio.NewScanner(file)
	//	for scanner.Scan() {
	//		fmt.Println(scanner.Text())
	//	}
	//
	//	if err := scanner.Err(); err != nil {
	//		log.Fatal(err)
	//	}
	//
}

//
//func perf2Seconds() {
//	// initialize the time to 0
//	// 	cat $name |\
//	//
//	//awk '
//	//NR == 1 {
//	//	# this assume no leading comments
//	//	if ($1 == "#yyy-mm-dd") {
//	//		getline
//	//	}
//	//	date = $1
//	//	sub("\\.[0-9]*", "", $2)
//	//	time = $2
//	//	print "#date time latency xfertime thinktime bytes transactions"
//	//}
//	// /^#/ { echo $0; next } # This does comments: contradiction
//	// /.*/ {
//	//	sub("\\.[0-9]*", "", $2)
//	//	if (time != $2) {
//	//		report(date, time, latency, xfertime, thinktime,
//	//		bytes, transactions)
//	//		date = $1
//	//		time = $2
//	//		latency = $3
//	//		xfertime = $4
//	//		thinktime = $5
//	//		bytes = $6
//	//		transactions = 0
//	//	}
//	//	else {
//	//		latency += $3
//	//		xfertime += $4
//	//		thinktime += $5
//	//		bytes += $6
//	//		transactions++
//	//	}
//	//}
//	//END {
//	//	report(date, time, latency, xfertime, thinktime, bytes, transacyions)
//	//}
//}
//
////func report(date, time, latency, xfertime, thinktime, bytes, xacts) {
////	if xacts > 0 {
////		printf("%s %s %f %f %f %d %d\n",
////			date, time, latency/xacts, xfertime/xacts,
////			thinktime/xacts, bytes/xacts, xacts)
////	}
