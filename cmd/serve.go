package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PextraCloud/pce-mcp/internal/config"
	"github.com/PextraCloud/pce-mcp/internal/server"
	"github.com/PextraCloud/pce-mcp/pkg/api"
	"github.com/spf13/cobra"
)

var (
	flagSSEAddr        string
	flagHTTPAddr       string
	flagPCEBaseURL     string
	flagInsecureTLS    bool
	flagTimeoutSeconds int
)

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVar(&flagSSEAddr, "sse-addr", ":2222", fmt.Sprintf("SSE server listen address, set to empty string to disable, overridable via %s env var", config.EnvSSEAddr))
	serveCmd.Flags().StringVar(&flagHTTPAddr, "http-addr", ":2223", fmt.Sprintf("HTTP server listen address, set to empty string to disable, overridable via %s env var", config.EnvHTTPAddr))
	serveCmd.Flags().StringVar(&flagPCEBaseURL, "base-url", "", fmt.Sprintf("Pextra CloudEnvironment(R) base URL (e.g., https://192.168.1.27:5007), overridable via %s env var", config.EnvBaseURL))
	serveCmd.Flags().BoolVar(&flagInsecureTLS, "tls-skip-verify", false, fmt.Sprintf("Skip TLS certificate verification for Pextra CloudEnvironment(R) API client. This may make you vulnerable to man-in-the-middle attacks; overridable via %s env var", config.EnvTLSSkipVerify))
	serveCmd.Flags().IntVar(&flagTimeoutSeconds, "timeout", 10, fmt.Sprintf("Timeout in seconds for Pextra CloudEnvironment(R) API client requests, overridable via %s env var", config.EnvTimeout))
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Build config with env fallbacks
		c, err := config.WithEnvDefaults(config.AppConfig{
			SSEAddr:           flagSSEAddr,
			HTTPAddr:          flagHTTPAddr,
			PCEBaseURL:        flagPCEBaseURL,
			PCEInsecureTLS:    flagInsecureTLS,
			PCEDefaultTimeout: time.Duration(flagTimeoutSeconds) * time.Second,
		})
		if err != nil {
			return err
		}
		config.Set(*c)

		// Construct client to validate config
		if _, err := api.NewClient(c.PCEBaseURL, c.PCEInsecureTLS, c.PCEDefaultTimeout); err != nil {
			return err
		}

		s := server.GetServer()
		server.AddTools(s)

		// Start servers (empty address disables per-flag help)
		errCh := make(chan error, 2)
		started := 0
		if c.SSEAddr != "" {
			log.Printf("Serving SSE at %s", c.SSEAddr)
			started++
			go func() { errCh <- server.StartSSE(s, c.SSEAddr) }()
		} else {
			log.Printf("SSE server disabled (empty address)")
		}
		if c.HTTPAddr != "" {
			log.Printf("Serving HTTP at %s", c.HTTPAddr)
			started++
			go func() { errCh <- server.StartStreamableHTTP(s, c.HTTPAddr) }()
		} else {
			log.Printf("HTTP server disabled (empty address)")
		}
		if started == 0 {
			return fmt.Errorf("both servers disabled; provide --sse-addr and/or --http-addr or env overrides")
		}

		// Graceful shutdown on signal
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-sigCh:
			log.Printf("shutdown signal received: %v", sig)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			<-ctx.Done()
			return nil
		case err := <-errCh:
			if err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		}
	},
}
