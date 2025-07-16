package main

import (
	"log/slog"
	"math"
	"net"
	"sync"
	"time"

	"github.com/sandertv/go-raknet"
)

func GetLowestPingIP(addr string, port string, log *slog.Logger) string {
	addrs, err := net.LookupIP(addr)
	if err != nil {
		return addr
	}
	var lowestPing time.Duration = -1
	var bestAddr net.IP
	var g sync.WaitGroup
	g.Add(len(addrs))
	for _, i := range addrs {
		currentAddr := i
		go func() {
			defer g.Done()
			ping := getAverageRaknetPing(net.JoinHostPort(currentAddr.String(), port), 3)
			log.Info("Pinged IP", "ip", currentAddr.String(), "ping", ping.Milliseconds())
			if lowestPing == -1 || ping < lowestPing {
				lowestPing = ping
				bestAddr = currentAddr
			}
		}()
	}
	g.Wait()
	if bestAddr == nil {
		return addr
	}
	return bestAddr.String()
}

func getAverageRaknetPing(addr string, numPings int) time.Duration {
	var totalPing time.Duration
	var numSuccessfulPings int
	for i := 0; i < numPings; i++ {
		startTime := time.Now()
		if _, err := raknet.Ping(addr); err == nil {
			totalPing += time.Since(startTime)
			numSuccessfulPings++
		}

	}
	if numSuccessfulPings == 0 {
		return math.MaxInt64
	}
	return totalPing / time.Duration(numSuccessfulPings)
}
