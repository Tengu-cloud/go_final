package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	url          = "http://127.0.0.1:80/_stats"
	pollInterval = 1 * time.Second
	maxErrCount  = 3
)

func main() {
	client := &http.Client{Timeout: 5 * time.Second}
	errCount := 0

	for {
		err := fetchAndProcess(client)
		if err != nil {
			errCount++
			if errCount >= maxErrCount {
				fmt.Println("Unable to fetch server statistic.")
			}
		} else {
			errCount = 0
		}
		time.Sleep(pollInterval)
	}
}

func fetchAndProcess(client *http.Client) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	parts := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(parts) != 7 {
		return fmt.Errorf("invalid stats format")
	}

	load := parseInt(parts[0])
	memTotal := parseInt(parts[1])
	memUsed := parseInt(parts[2])
	diskTotal := parseInt(parts[3])
	diskUsed := parseInt(parts[4])
	netTotal := parseInt(parts[5])
	netUsed := parseInt(parts[6])

	// 🔹 Network
	if netUsed*100/netTotal > 90 {
		netLeft := (netTotal - netUsed) / 1_000_000
		fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", netLeft)
	}

	// 🔹 Load Average
	if load > 30 {
		fmt.Printf("Load Average is too high: %d\n", load)
	}

	// 🔹 Memory
	memPercent := int(float32(memUsed) / float32(memTotal) * 100)
	if memPercent > 80 {
		fmt.Printf("Memory usage too high: %d%%\n", memPercent)
	}

	// 🔹 Disk
	if diskUsed*100/diskTotal > 90 {
		diskLeft := (diskTotal - diskUsed) / (1024 * 1024) // MiB
		fmt.Printf("Free disk space is too low: %d Mb left\n", diskLeft)
	}

	return nil
}

func parseInt(s string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return n
}
