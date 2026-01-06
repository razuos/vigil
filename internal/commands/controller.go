package commands

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"vigil/internal/controller/store"
	"vigil/internal/controller/telegram"
	"vigil/internal/controller/wol"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	state *store.Store
	bot   *telegram.Bot
)

// Prometheus Collector
type VigilCollector struct {
	store  *store.Store
	load   *prometheus.Desc
	status *prometheus.Desc
}

func NewVigilCollector(s *store.Store) *VigilCollector {
	return &VigilCollector{
		store:  s,
		load:   prometheus.NewDesc("vigil_server_load", "Current Load Average of Unraid Server", nil, nil),
		status: prometheus.NewDesc("vigil_server_status", "Unraid Server Status (1=Online, 0=Offline)", nil, nil),
	}
}

func (c *VigilCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.load
	ch <- c.status
}

func (c *VigilCollector) Collect(ch chan<- prometheus.Metric) {
	s := c.store.GetStatus()

	statusVal := 0.0
	if s.IsOnline {
		statusVal = 1.0
	}

	ch <- prometheus.MustNewConstMetric(c.load, prometheus.GaugeValue, s.Load)
	ch <- prometheus.MustNewConstMetric(c.status, prometheus.GaugeValue, statusVal)
}

// Metric Counters
var (
	wakeCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "vigil_wake_commands_total",
		Help: "Total number of Wake-on-LAN commands sent",
	})
	shutdownCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "vigil_shutdown_commands_total",
		Help: "Total number of Force Shutdown commands sent",
	})
)

func init() {
	// Register Counters globally
	prometheus.MustRegister(wakeCounter)
	prometheus.MustRegister(shutdownCounter)
}

func NewControllerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "controller",
		Short: "Run Vigil Controller",
		Run: func(cmd *cobra.Command, args []string) {
			startController()
		},
	}

	cmd.Flags().String("port", "8080", "HTTP Port")
	cmd.Flags().String("mac", "", "MAC Address of Unraid Server")
	cmd.Flags().String("tg-token", "", "Telegram Bot Token")
	cmd.Flags().String("tg-chat", "", "Telegram Chat ID")
	cmd.Flags().Float64("load-threshold", 1.5, "Load Average Threshold to consider idle")

	viper.BindPFlag("port", cmd.Flags().Lookup("port"))
	viper.BindPFlag("mac_address", cmd.Flags().Lookup("mac"))
	viper.BindPFlag("telegram_bot_token", cmd.Flags().Lookup("tg-token"))
	viper.BindPFlag("telegram_chat_id", cmd.Flags().Lookup("tg-chat"))
	viper.BindPFlag("load_threshold", cmd.Flags().Lookup("load-threshold"))

	return cmd
}

func startController() {
	port := viper.GetString("port")
	mac := viper.GetString("mac_address")
	tgToken := viper.GetString("telegram_bot_token")
	tgChat := viper.GetString("telegram_chat_id")
	loadThreshold := viper.GetFloat64("load_threshold")

	slog.Info("Starting Vigil Controller", "port", port, "chat_id", tgChat, "load_threshold", loadThreshold)

	state = store.NewStore()
	bot = telegram.NewBot(tgToken, tgChat)

	// Register Prometheus Collector (Gauge)
	prometheus.MustRegister(NewVigilCollector(state))

	// Register Telegram Commands
	bot.RegisterCommand("/start", func(args string) string {
		return "👋 Vigil is online. Use /status, /wake, /shutdown, or /keepawake."
	})

	bot.RegisterCommand("/status", func(args string) string {
		s := state.GetStatus()
		status := "🔴 Offline"
		if s.IsOnline {
			status = "🟢 Online"
		}
		return fmt.Sprintf("Status: %s\nLoad: %.2f\nOverride: %v", status, s.Load, s.Override)
	})

	bot.RegisterCommand("/wake", func(args string) string {
		if mac == "" {
			return "⚠️ WOL not configured (No MAC Address)."
		}
		if err := wol.SendMagicPacket(mac); err != nil {
			return fmt.Sprintf("❌ Failed to send WOL: %v", err)
		}
		wakeCounter.Inc()
		return "⚡ Magic Packet sent!"
	})

	bot.RegisterCommand("/shutdown", func(args string) string {
		state.RequestShutdown()
		shutdownCounter.Inc()
		return "🛑 Shutdown Request Queued. Agent should pick it up momentarily."
	})

	bot.RegisterCommand("/keepawake", func(args string) string {
		if args == "on" {
			state.SetOverride(true)
			return "🔴 Override ENABLED. Server will stay awake."
		} else if args == "off" {
			state.SetOverride(false)
			return "⚪ Override DISABLED. Auto-sleep active."
		}
		s := state.GetStatus()
		return fmt.Sprintf("Usage: /keepawake [on|off]. Current: %v", s.Override)
	})

	// Start Bot Polling
	bot.Start()
	defer bot.Stop()

	// API Routes (for Agent)
	http.HandleFunc("/api/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		handleHeartbeat(w, r, loadThreshold)
	})

	// Metrics Endpoint
	http.Handle("/metrics", promhttp.Handler())

	slog.Info("Server listening", "address", ":"+port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

type HeartbeatRequest struct {
	Load float64 `json:"load"`
}

type HeartbeatResponse struct {
	Action string `json:"action"`
}

func handleHeartbeat(w http.ResponseWriter, r *http.Request, threshold float64) {
	var req HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	prevStatus := state.GetStatus()

	action := "stay_awake"
	status := state.GetStatus()

	// Priority 1: Force Shutdown
	if state.ConsumeShutdownRequest() {
		action = "sleep"
		slog.Info("Force Shutdown Executed", "client_ip", r.RemoteAddr)
	} else if !status.Override && req.Load < threshold {
		// Priority 2: Normal Load Check
		action = "sleep"
	}

	state.UpdateHeartbeat(req.Load)

	// Notifications
	if !prevStatus.IsOnline {
		bot.Send("🟢 Vigil: Server is back Online!")
		slog.Info("Server Back Online", "load", req.Load)
	}

	json.NewEncoder(w).Encode(HeartbeatResponse{Action: action})
}
