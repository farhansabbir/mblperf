package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

func main() {
	filename := "api-performance." + strconv.Itoa(time.Now().Day()) + strconv.Itoa(int(time.Now().Month())) + strconv.Itoa(int(time.Now().Hour())) + strconv.Itoa(int(time.Now().Minute())) + ".csv"

	iterations := flag.Int("iterations", 0, "Number of iterations. Default 0 is infinite")
	sleep := flag.Int("sleep", 15, "Seconds to sleep between requests")
	flag.Parse()
	if len(flag.Args()) != 1 {
		fmt.Printf("Usage: %v [options] <url>\n", os.Args[0])
		fmt.Println("options")
		flag.PrintDefaults()
		os.Exit(1)
	}

	URL, err := url.Parse(flag.Args()[0])
	if err != nil {
		fmt.Printf("Invalid URL %v\n", err)
		os.Exit(1)
	}
	remainingIterations := *iterations
	fmt.Println("Output will be written to: " + filename)
	fmt.Println("DateTime, Time_To_DNS_Lookup, IPs, Response_Code, Time_To_Response_Code, Size_Of_Body, Time_To_Read_Body")
	csvfile, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating csv file: %v\n", err)
		os.Exit(1)
	}
	csvfile.WriteString("DateTime, Time_To_DNS_Lookup, IPs, Response_Code, Time_To_Response_Code, Size_Of_Body, Time_To_Read_Body\n")
	for {
		var statusLine bytes.Buffer
		statusLine.WriteString(time.Now().Format("2006-Jan-02 15:04:05, "))
		start := time.Now().UnixMilli()
		ips, err := net.LookupIP(URL.Hostname())
		end := time.Now().UnixMilli()

		statusLine.WriteString(fmt.Sprintf("%v, ", (end - start))) // DNS
		if err != nil {
			fmt.Printf("Error getting IP Address from URL: %v\n", err)
			os.Exit(1)
		}

		statusLine.WriteString(fmt.Sprintf("%v, ", ips))

		client := http.Client{Timeout: time.Duration(90) * time.Second}
		start = time.Now().UnixMilli()
		resp, err := client.Get(URL.String())
		if err != nil {
			fmt.Printf("An error occured while connecting to URL. %v", err)
			os.Exit(1)
		}

		status := resp.StatusCode
		end = time.Now().UnixMilli()
		statusLine.WriteString(fmt.Sprintf("%v, %v, ", status, (end - start))) // resp code and url resp time

		start = time.Now().UnixMilli()
		body, err := ioutil.ReadAll(resp.Body)
		end = time.Now().UnixMilli()
		if err != nil {
			fmt.Println("An error occured while connecting to URL")
			os.Exit(1)
		}
		statusLine.WriteString(fmt.Sprintf("%v, %v\n", len(body), (end - start))) // len of body and body read time
		resp.Body.Close()
		if *iterations != 0 || remainingIterations > 0 {
			remainingIterations -= 1
			if remainingIterations == 0 {
				break
			}
		}
		fmt.Printf(statusLine.String())
		csvfile.WriteString(statusLine.String())
		csvfile.Sync()
		defer csvfile.Close()
		time.Sleep(time.Duration(*sleep) * time.Second)
	}
}
