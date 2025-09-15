package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type PendingData struct {
	ServerID    int     `json:"server_id"`
	Load        string  `json:"load"`
	Cpu         float64 `json:"cpu"`
	Ram         float64 `json:"ram"`
	Disk        float64 `json:"disk"`
	NetIn       uint64  `json:"net_in"`
	NetOut      uint64  `json:"net_out"`
	CollectedAt string  `json:"collected_at"`
}

func main() {
	cfg := getConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", cfg.MySQL.Username, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.DBName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxIdleConns(0)

	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Fatal(err)
	}

	pendentes := loadCache()
	fmt.Println("Monitor Running")
	cpu.Percent(time.Second, false)

	for {
		collectAndInsert(db, cfg.ServerID, loc, &pendentes)
		time.Sleep(3 * time.Minute)
	}
}

func collectAndInsert(db *sql.DB, serverID int, loc *time.Location, pendentes *[]PendingData) {
	for i := 0; i < len(*pendentes); i++ {
		p := (*pendentes)[i]
		_, err := db.Exec("INSERT INTO server_stats (server_id, server_running_state, cpu_usage, ram_usage, disk_usage, network_in, network_out, collected_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			p.ServerID, p.Load, p.Cpu, p.Ram, p.Disk, p.NetIn, p.NetOut, p.CollectedAt)
		if err != nil {
			log.Println("Failed to send pending data:", err)
			break
		}
		*pendentes = append((*pendentes)[:i], (*pendentes)[i+1:]...)
		i--
	}
	saveCache(*pendentes)

	cpuUsage, _ := cpu.Percent(0, false)
	ramUsage, _ := mem.VirtualMemory()
	diskUsage, _ := disk.Usage("/")
	netStats, _ := net.IOCounters(false)

	var networkIn, networkOut uint64
	if len(netStats) > 0 {
		networkIn = netStats[0].BytesRecv
		networkOut = netStats[0].BytesSent
	}

	load := estimateLoad(cpuUsage[0], ramUsage.UsedPercent, diskUsage.UsedPercent)

	collectedAt := time.Now().In(loc).Format("2006-01-02 15:04:05")

	_, err := db.Exec("INSERT INTO server_stats (server_id, server_running_state, cpu_usage, ram_usage, disk_usage, network_in, network_out, collected_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		serverID, load, cpuUsage[0], ramUsage.UsedPercent, diskUsage.UsedPercent, networkIn, networkOut, collectedAt)
	if err != nil {
		log.Println("Failed to insert new data:", err)
		*pendentes = append(*pendentes, PendingData{serverID, load, cpuUsage[0], ramUsage.UsedPercent, diskUsage.UsedPercent, networkIn, networkOut, collectedAt})
		saveCache(*pendentes)
	}
}

func loadCache() []PendingData {
	file, err := os.Open("cache.json")
	if err != nil {
		return []PendingData{}
	}
	defer file.Close()
	var pendentes []PendingData
	json.NewDecoder(file).Decode(&pendentes)
	return pendentes
}

func saveCache(pendentes []PendingData) {
	file, _ := os.Create("cache.json")
	defer file.Close()
	json.NewEncoder(file).Encode(pendentes)
}

func estimateLoad(cpu, ram, disk float64) string {
	if cpu > 90 || ram > 90 {
		return "critical"
	}
	if cpu > 60 || ram > 60 || disk > 80 {
		return "high"
	}
	if cpu > 40 || ram > 40 {
		return "medium"
	}
	return "low"
}
