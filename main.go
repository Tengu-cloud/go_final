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
	url           = "http://srv.msk01.gigacorp.local/_stats"
	pollInterval  = 10 * time.Second
	maxErrorCount = 3
)

func main() {
	errorCount := 0
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for {
		err := fetchAndProcessStats(client)
		if err != nil {
			errorCount++
			if errorCount >= maxErrorCount {
				fmt.Println("Unable to fetch server statistic.")
			}
		} else {
			errorCount = 0
		}

		time.Sleep(pollInterval)
	}
}

func fetchAndProcessStats(client *http.Client) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 status")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	parts := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(parts) != 7 {
		return fmt.Errorf("invalid stats format")
	}

	values := make([]float64, 7)
	for i, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return err
		}
		values[i] = v
	}

	loadAvg := values[0]
	memTotal := values[1]
	memUsed := values[2]
	diskTotal := values[3]
	diskUsed := values[4]
	netTotal := values[5]
	netUsed := values[6]

	// Load Average
	if loadAvg > 30 {
		fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
	}

	// Memory usage
	memUsagePercent := (memUsed / memTotal) * 100
	if memUsagePercent > 80 {
		fmt.Printf("Memory usage too high: %.0f%%\n", memUsagePercent)
	}

	// Disk space
	freeDiskBytes := diskTotal - diskUsed
	freeDiskMB := freeDiskBytes / (1024 * 1024)
	diskUsagePercent := (diskUsed / diskTotal) * 100
	if diskUsagePercent > 90 {
		fmt.Printf("Free disk space is too low: %.0f Mb left\n", freeDiskMB)
	}

	// Network bandwidth
	netFreeBytes := netTotal - netUsed
	netFreeMbit := (netFreeBytes * 8) / (1024 * 1024)
	if netUsed > netTotal*0.9 {
		fmt.Printf(
			"Network bandwidth usage high: %.0f Mbit/s available\n",
			netFreeMbit,
		)
	}

	return nil
}
