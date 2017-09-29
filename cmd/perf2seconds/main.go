// perf2Seconds reduces a perf .csv file to a by-seconds perf .csv file
package main

//import (
//	"bufio"
//	"fmt"
//	"io"
//	"log"
//	"os"
//	"os/exec"
//)
//
//func main() {
//	var err error
//	var stdin, file io.WriteCloser
//
//	if len(os.Args) == 0 {
//		log.Fatalf("Usage: %s -|file", os.Args[0])
//	}
//
//	cmdName := "sort"
//	cmdArgs := []string{"-k2.1,2.2nb", "-k2.5,2.6nb", "-k2.8,2.12nb",
//		"--temporary-directory=/var/tmp"}
//	cmd := exec.Command(cmdName, cmdArgs...)
//	if os.Args[1] == "-" {
//		stdin, err = cmd.StdinPipe()
//		if err != nil {
//			log.Fatalf("%v: unable to open stdin, %v\n",
//				os.Args[0], err)
//		}
//		defer stdin.Close() // nolint
//	} else {
//		file, err = os.Open(os.Args[1])
//		if err != nil {
//			log.Fatalf("%v: could not open %q, %v\n",
//				os.Args[0], os.Args[1], err)
//		}
//		defer file.Close()
//		cmd.Stdin = file
//	}
//
//	out, err := cmd.CombinedOutput()
//	fmt.Printf("%s\n", out)
//	if err != nil {
//		log.Fatalf("%v: unable to run cat, %v\n", os.Args[0], err)
//	}
//	//cmdReader, err := cmd.StdoutPipe()
//	//if err != nil {
//	//	fmt.Errorf("Error creating StdoutPipe for Cmd, %v\n", err)
//	//}
//	//scanner := bufio.NewScanner(cmdReader)
//	//
//	//err = cmd.Run()
//	//if err != nil {
//	//		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
//	//		os.Exit(1)
//	//}
//	//perf2Seconds(scanner)
//}
//
//// nolint
//func perf2Seconds(scanner *bufio.Scanner) {
//
//	for scanner.Scan() {
//		fmt.Printf("Got %s\n", scanner.Text())
//	}
//}

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

//next bit:
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
//}

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
