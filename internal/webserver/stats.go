package webserver

import (
	"encoding/json"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/blubskye/himiko/internal/database"
	"github.com/bwmarrin/discordgo"
)

// MetricsSnapshot represents a point-in-time measurement
type MetricsSnapshot struct {
	Timestamp time.Time `json:"timestamp"`

	// Memory stats (from runtime.MemStats)
	MemAlloc      uint64 `json:"mem_alloc"`       // Bytes currently allocated
	MemTotalAlloc uint64 `json:"mem_total_alloc"` // Cumulative bytes allocated
	MemSys        uint64 `json:"mem_sys"`         // Total memory from OS
	MemNumGC      uint32 `json:"mem_num_gc"`      // Number of GC runs
	Goroutines    int    `json:"goroutines"`      // Number of goroutines

	// Discord stats
	GuildCount       int   `json:"guild_count"`
	MemberCount      int   `json:"member_count"`
	ChannelCount     int   `json:"channel_count"`
	HeartbeatLatency int64 `json:"heartbeat_latency_ms"` // Discord WS latency in ms

	// Activity stats (counters)
	CommandsTotal int64 `json:"commands_total"` // Total commands processed
	MessagesTotal int64 `json:"messages_total"` // Total messages seen

	// Rate stats (per-interval calculations)
	CommandsPerMin float64 `json:"commands_per_min"`
	MessagesPerMin float64 `json:"messages_per_min"`
}

// DatabaseStats holds SQLite database statistics
type DatabaseStats struct {
	FileSizeBytes int64            `json:"file_size_bytes"`
	TableCounts   map[string]int64 `json:"table_counts"`
}

// RealTimeStats holds all real-time statistics data
type RealTimeStats struct {
	Current       *MetricsSnapshot `json:"current"`
	StartTime     time.Time        `json:"start_time"`
	UptimeSeconds int64            `json:"uptime_seconds"`
	Database      *DatabaseStats   `json:"database,omitempty"`
	Version       string           `json:"version"`
}

// RingBuffer is a fixed-size circular buffer for historical metrics
type RingBuffer struct {
	mu    sync.RWMutex
	data  []MetricsSnapshot
	size  int
	head  int // Next write position
	count int // Current number of elements
}

// NewRingBuffer creates a new ring buffer with the specified capacity
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]MetricsSnapshot, size),
		size: size,
	}
}

// Push adds a new snapshot to the buffer
func (rb *RingBuffer) Push(snapshot MetricsSnapshot) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.data[rb.head] = snapshot
	rb.head = (rb.head + 1) % rb.size
	if rb.count < rb.size {
		rb.count++
	}
}

// GetAll returns all snapshots in chronological order
func (rb *RingBuffer) GetAll() []MetricsSnapshot {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return nil
	}

	result := make([]MetricsSnapshot, rb.count)
	start := (rb.head - rb.count + rb.size) % rb.size
	for i := 0; i < rb.count; i++ {
		result[i] = rb.data[(start+i)%rb.size]
	}
	return result
}

// GetLast returns the most recent n snapshots
func (rb *RingBuffer) GetLast(n int) []MetricsSnapshot {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return nil
	}

	count := n
	if count > rb.count {
		count = rb.count
	}

	result := make([]MetricsSnapshot, count)
	start := (rb.head - count + rb.size) % rb.size
	for i := 0; i < count; i++ {
		result[i] = rb.data[(start+i)%rb.size]
	}
	return result
}

// GetLatest returns the most recent snapshot
func (rb *RingBuffer) GetLatest() *MetricsSnapshot {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return nil
	}

	idx := (rb.head - 1 + rb.size) % rb.size
	snapshot := rb.data[idx]
	return &snapshot
}

// StatsCollector manages metrics collection and history
type StatsCollector struct {
	mu        sync.RWMutex
	session   *discordgo.Session
	db        *database.DB
	dbPath    string
	startTime time.Time
	version   string

	// Ring buffers for different time scales
	// 5-second intervals for last hour = 720 samples
	hourlyHistory *RingBuffer
	// 1-minute intervals for last day = 1440 samples
	dailyHistory *RingBuffer

	// Counters for rate calculations
	commandCount     int64
	messageCount     int64
	lastCommandCount int64
	lastMessageCount int64
	lastCountTime    time.Time

	// SSE clients
	clients   map[chan []byte]struct{}
	clientsMu sync.RWMutex

	stopChan chan struct{}
	running  bool
}

// NewStatsCollector creates a new stats collector
func NewStatsCollector(session *discordgo.Session, db *database.DB, dbPath string, startTime time.Time, version string) *StatsCollector {
	return &StatsCollector{
		session:       session,
		db:            db,
		dbPath:        dbPath,
		startTime:     startTime,
		version:       version,
		hourlyHistory: NewRingBuffer(720),  // 5-sec intervals * 720 = 1 hour
		dailyHistory:  NewRingBuffer(1440), // 1-min intervals * 1440 = 1 day
		clients:       make(map[chan []byte]struct{}),
		stopChan:      make(chan struct{}),
		lastCountTime: time.Now(),
	}
}

// Start begins the metrics collection goroutine
func (sc *StatsCollector) Start() {
	sc.mu.Lock()
	if sc.running {
		sc.mu.Unlock()
		return
	}
	sc.running = true
	sc.stopChan = make(chan struct{})
	sc.mu.Unlock()

	go sc.collectLoop()
}

// Stop stops the metrics collection
func (sc *StatsCollector) Stop() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if !sc.running {
		return
	}
	sc.running = false
	close(sc.stopChan)
}

// IsRunning returns whether the collector is running
func (sc *StatsCollector) IsRunning() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.running
}

// collectLoop runs the main collection loop
func (sc *StatsCollector) collectLoop() {
	fastTicker := time.NewTicker(5 * time.Second) // For hourly history & SSE
	slowTicker := time.NewTicker(1 * time.Minute) // For daily history
	defer fastTicker.Stop()
	defer slowTicker.Stop()

	// Collect initial snapshot
	snapshot := sc.collectSnapshot()
	sc.hourlyHistory.Push(snapshot)
	sc.dailyHistory.Push(snapshot)

	for {
		select {
		case <-sc.stopChan:
			return
		case <-fastTicker.C:
			snapshot := sc.collectSnapshot()
			sc.hourlyHistory.Push(snapshot)
			sc.broadcastToClients(snapshot)
		case <-slowTicker.C:
			snapshot := sc.collectSnapshot()
			sc.dailyHistory.Push(snapshot)
		}
	}
}

// collectSnapshot gathers all current metrics
func (sc *StatsCollector) collectSnapshot() MetricsSnapshot {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Calculate Discord stats
	guildCount := 0
	memberCount := 0
	channelCount := 0

	if sc.session != nil && sc.session.State != nil {
		sc.session.State.RLock()
		guilds := sc.session.State.Guilds
		guildCount = len(guilds)
		for _, g := range guilds {
			memberCount += g.MemberCount
			channelCount += len(g.Channels)
		}
		sc.session.State.RUnlock()
	}

	// Get heartbeat latency
	var latency int64
	if sc.session != nil {
		latency = sc.session.HeartbeatLatency().Milliseconds()
	}

	// Calculate rates
	sc.mu.Lock()
	elapsed := time.Since(sc.lastCountTime).Minutes()
	if elapsed < 0.01 {
		elapsed = 0.01 // Prevent division by zero
	}
	cmdRate := float64(sc.commandCount-sc.lastCommandCount) / elapsed
	msgRate := float64(sc.messageCount-sc.lastMessageCount) / elapsed
	sc.lastCommandCount = sc.commandCount
	sc.lastMessageCount = sc.messageCount
	sc.lastCountTime = time.Now()
	cmdTotal := sc.commandCount
	msgTotal := sc.messageCount
	sc.mu.Unlock()

	return MetricsSnapshot{
		Timestamp:        time.Now(),
		MemAlloc:         memStats.Alloc,
		MemTotalAlloc:    memStats.TotalAlloc,
		MemSys:           memStats.Sys,
		MemNumGC:         memStats.NumGC,
		Goroutines:       runtime.NumGoroutine(),
		GuildCount:       guildCount,
		MemberCount:      memberCount,
		ChannelCount:     channelCount,
		HeartbeatLatency: latency,
		CommandsTotal:    cmdTotal,
		MessagesTotal:    msgTotal,
		CommandsPerMin:   cmdRate,
		MessagesPerMin:   msgRate,
	}
}

// IncrementCommand increments the command counter
func (sc *StatsCollector) IncrementCommand() {
	sc.mu.Lock()
	sc.commandCount++
	sc.mu.Unlock()
}

// IncrementMessage increments the message counter
func (sc *StatsCollector) IncrementMessage() {
	sc.mu.Lock()
	sc.messageCount++
	sc.mu.Unlock()
}

// GetCurrentSnapshot returns the current metrics snapshot
func (sc *StatsCollector) GetCurrentSnapshot() *MetricsSnapshot {
	snapshot := sc.collectSnapshot()
	return &snapshot
}

// GetRealTimeStats returns all real-time statistics
func (sc *StatsCollector) GetRealTimeStats() *RealTimeStats {
	snapshot := sc.collectSnapshot()
	return &RealTimeStats{
		Current:       &snapshot,
		StartTime:     sc.startTime,
		UptimeSeconds: int64(time.Since(sc.startTime).Seconds()),
		Version:       sc.version,
	}
}

// GetHourlyHistory returns the hourly history snapshots
func (sc *StatsCollector) GetHourlyHistory() []MetricsSnapshot {
	return sc.hourlyHistory.GetAll()
}

// GetDailyHistory returns the daily history snapshots
func (sc *StatsCollector) GetDailyHistory() []MetricsSnapshot {
	return sc.dailyHistory.GetAll()
}

// GetDatabaseStats collects database file and table statistics
func (sc *StatsCollector) GetDatabaseStats() *DatabaseStats {
	stats := &DatabaseStats{
		TableCounts: make(map[string]int64),
	}

	// Get database file size
	if sc.dbPath != "" {
		if fileInfo, err := os.Stat(sc.dbPath); err == nil {
			stats.FileSizeBytes = fileInfo.Size()
		}
	}

	// Get table counts for key tables
	if sc.db != nil {
		tables := []string{
			"guild_settings", "command_history", "warnings", "user_xp",
			"mod_actions", "user_activity", "deleted_messages", "custom_commands",
		}
		for _, table := range tables {
			var count int64
			row := sc.db.QueryRow("SELECT COUNT(*) FROM " + table)
			if row.Scan(&count) == nil {
				stats.TableCounts[table] = count
			}
		}
	}

	return stats
}

// RegisterClient adds a client channel for SSE updates
func (sc *StatsCollector) RegisterClient(ch chan []byte) {
	sc.clientsMu.Lock()
	sc.clients[ch] = struct{}{}
	sc.clientsMu.Unlock()
}

// UnregisterClient removes a client channel
func (sc *StatsCollector) UnregisterClient(ch chan []byte) {
	sc.clientsMu.Lock()
	delete(sc.clients, ch)
	sc.clientsMu.Unlock()
}

// broadcastToClients sends update to all SSE clients
func (sc *StatsCollector) broadcastToClients(snapshot MetricsSnapshot) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return
	}

	sc.clientsMu.RLock()
	defer sc.clientsMu.RUnlock()

	for ch := range sc.clients {
		select {
		case ch <- data:
		default:
			// Client buffer full, skip
		}
	}
}

// GetStartTime returns the bot start time
func (sc *StatsCollector) GetStartTime() time.Time {
	return sc.startTime
}

// GetVersion returns the bot version
func (sc *StatsCollector) GetVersion() string {
	return sc.version
}
