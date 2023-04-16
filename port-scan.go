package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

type PortScanner struct {
	ip   string
	lock *semaphore.Weighted
}

var target string

func Ulimit() int64 {
	out, err := exec.Command("ulimit", "-n").Output()

	if err != nil {
		panic(err)
	}

	s := strings.TrimSpace(string(out))
	i, err := strconv.ParseInt(s, 10, 64)

	if err != nil {
		panic(err)
	}

	return i
}

func ScanPort(ip string, port int, timeout time.Duration) {
	target := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", target, timeout)

	if err != nil {
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			ScanPort(ip, port, timeout)
		} else {
			fmt.Println(port, "CLOSED")
		}
		return
	}

	conn.Close()
	fmt.Println(port, "OPEN")
}

func (ps *PortScanner) Start(from int, to int, timeout time.Duration) {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	for port := from; port <= to; port++ {
		wg.Add(1)
		ps.lock.Acquire(context.TODO(), 1)

		go func(port int) {
			defer ps.lock.Release(1)
			defer wg.Done()
			ScanPort(ps.ip, port, timeout)
		}(port)
	}
}

func init() {
	flag.StringVar(&target, "target", "127.0.0.1", "IP address or hostname of target")
	flag.Parse()
}

func main() {
	fmt.Println("Performing port scan for " + target + "...")

	ps := &PortScanner{
		ip:   target,
		lock: semaphore.NewWeighted(Ulimit()),
	}

	ps.Start(1, 65535, 500*time.Millisecond)
}
