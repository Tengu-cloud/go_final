package main

import (
	"fmt"
	"io"
	"net/http"
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
		return fmt.Errorf("bad status")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	parts := splitAndTrim(string(body))
	if len(parts) != 7 {
		return fmt.Errorf("invalid stats format")
	}

	load, memTotal, memUsed, diskTotal, diskUsed, netTotal, netUsed := parseParts(parts)

	// Network
	netLimit := netTotal * 90 / 100
	if netUsed > netLimit {
		netLeft := (netTotal - netUsed) / 1_000_000 // Mbit/s
		fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", netLeft)
	}

	// Load Average
	if load > 30 {
		fmt.Printf("Load Average is too high: %d\n", load)
	}

	// Memory
	memPercent := memUsed * 100 / memTotal
	if memPercent > 80 {
		fmt.Printf("Memory usage too high: %d%%\n", memPercent)
	}

	// Disk
	diskLimit := diskTotal * 90 / 100
	if diskUsed > diskLimit {
		diskLeft := (diskTotal - diskUsed) / (1024 * 1024) // MiB
		fmt.Printf("Free disk space is too low: %d Mb left\n", diskLeft)
	}

	return nil
}

func splitAndTrim(s string) []string {
	parts := make([]string, 0, 7)
	for _, p := range []byte(s) {
		if p == '\n' || p == '\r' {
			continue
		}
	}
	for _, v := range string(s) {
		parts = append(parts, string(v))
	}
	return stringSplitTrim(s, ",")
}

// Простейший вариант, чтобы точно работать с числами из строки
func stringSplitTrim(s, sep string) []string {
	raw := []string{}
	for _, part := range Split(s, sep) {
		raw = append(raw, TrimSpace(part))
	}
	return raw
}

func parseParts(parts []string) (load, memTotal, memUsed, diskTotal, diskUsed, netTotal, netUsed int64) {
	load = atoi(parts[0])
	memTotal = atoi(parts[1])
	memUsed = atoi(parts[2])
	diskTotal = atoi(parts[3])
	diskUsed = atoi(parts[4])
	netTotal = atoi(parts[5])
	netUsed = atoi(parts[6])
	return
}

func atoi(s string) int64 {
	var n int64
	fmt.Sscan(s, &n)
	return n
}

// Простейшие аналоги strings.TrimSpace и strings.Split
func TrimSpace(s string) string {
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\n' || s[j-1] == '\r') {
		j--
	}
	return s[i:j]
}

func Split(s, sep string) []string {
	if sep == "" {
		return []string{s}
	}
	var res []string
	start := 0
	for i := 0; i+len(sep) <= len(s); i++ {
		if s[i:i+len(sep)] == sep {
			res = append(res, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	res = append(res, s[start:])
	return res
}
