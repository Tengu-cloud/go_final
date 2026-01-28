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
	url          = "http://srv.msk01.gigacorp.local/_stats"
	pollInterval = 10 * time.Second
	maxErrCount  = 3
)

func main() {
	client := &http.Client{Timeout: 5 * time.Second}
	errCount := 0

	for {
		if err := fetchAndProcess(client); err != nil {
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

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	parts := strings.Split(strings.TrimSpace(string(body)), ",")
	if len(parts) != 7 {
		return fmt.Errorf("bad format")
	}

	vals := make([]uint64, 7)
	for i, p := range parts {
		v, err := strconv.ParseUint(strings.TrimSpace(p), 10, 64)
		if err != nil {
			return err
		}
		vals[i] = v
	}

	load := vals[0]
	memTotal := vals[1]
	memUsed := vals[2]
	diskTotal := vals[3]
	diskUsed := vals[4]
	netTotal := vals[5]
	netUsed := vals[6]

	// Load Average
	if load > 30 {
		fmt.Printf("Load Average is too high: %d\n", load)
	}

	// Memory
	memPercent := memUsed * 100 / memTotal
	if memPercent > 80 {
		fmt.Printf("Memory usage too high: %d%%\n", memPercent)
	}

	// Disk (90% limit, decimal MB)
	diskLimit := diskTotal * 90 / 100
	if diskUsed > diskLimit {
		left := diskTotal - diskUsed
		mbLeft := left / 1_000_000
		fmt.Printf("Free disk space is too low: %d Mb left\n", mbLeft)
	}

	// Network (90% limit, decimal Mbit)
	netLimit := netTotal * 90 / 100
	if netUsed > netLimit {
		left := netTotal - netUsed
		mbitLeft := (left * 8) / 1_000_000
		fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", mbitLeft)
	}

	return nil
}
