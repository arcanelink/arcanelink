package discovery

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/arcane/arcanelink/pkg/logger"
	"github.com/arcane/arcanelink/pkg/models"
	"go.uber.org/zap"
)

type ServerResolver struct {
	cache      map[string]*cacheEntry
	cacheMutex sync.RWMutex
	httpClient *http.Client
}

type cacheEntry struct {
	serverInfo *models.ServerInfo
	expiresAt  time.Time
}

func NewServerResolver() *ServerResolver {
	return &ServerResolver{
		cache: make(map[string]*cacheEntry),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Resolve resolves a domain to server information
func (r *ServerResolver) Resolve(domain string) (*models.ServerInfo, error) {
	// Check cache first
	r.cacheMutex.RLock()
	if entry, ok := r.cache[domain]; ok {
		if time.Now().Before(entry.expiresAt) {
			r.cacheMutex.RUnlock()
			logger.Debug("Server info from cache", zap.String("domain", domain))
			return entry.serverInfo, nil
		}
	}
	r.cacheMutex.RUnlock()

	// Try .well-known first
	serverInfo, err := r.resolveWellKnown(domain)
	if err == nil {
		r.cacheServerInfo(domain, serverInfo)
		return serverInfo, nil
	}

	logger.Debug(".well-known resolution failed, trying DNS SRV", zap.String("domain", domain), zap.Error(err))

	// Try DNS SRV
	serverInfo, err = r.resolveDNSSRV(domain)
	if err == nil {
		r.cacheServerInfo(domain, serverInfo)
		return serverInfo, nil
	}

	logger.Debug("DNS SRV resolution failed, using fallback", zap.String("domain", domain), zap.Error(err))

	// Fallback to domain:8448
	serverInfo = &models.ServerInfo{
		Host: domain,
		Port: 8448,
	}
	r.cacheServerInfo(domain, serverInfo)
	return serverInfo, nil
}

// resolveWellKnown resolves server info via .well-known
func (r *ServerResolver) resolveWellKnown(domain string) (*models.ServerInfo, error) {
	url := fmt.Sprintf("https://%s/.well-known/matrix/server", domain)

	resp, err := r.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch .well-known: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(".well-known returned status %d", resp.StatusCode)
	}

	var wellKnown models.WellKnownResponse
	if err := json.NewDecoder(resp.Body).Decode(&wellKnown); err != nil {
		return nil, fmt.Errorf("failed to decode .well-known: %w", err)
	}

	// Parse server string (format: "host:port" or "host")
	host, port, err := net.SplitHostPort(wellKnown.Server)
	if err != nil {
		// No port specified, use default
		return &models.ServerInfo{
			Host: wellKnown.Server,
			Port: 8448,
		}, nil
	}

	portNum := 8448
	fmt.Sscanf(port, "%d", &portNum)

	return &models.ServerInfo{
		Host: host,
		Port: portNum,
	}, nil
}

// resolveDNSSRV resolves server info via DNS SRV record
func (r *ServerResolver) resolveDNSSRV(domain string) (*models.ServerInfo, error) {
	_, srvs, err := net.LookupSRV("matrix-fed", "tcp", domain)
	if err != nil {
		return nil, fmt.Errorf("DNS SRV lookup failed: %w", err)
	}

	if len(srvs) == 0 {
		return nil, fmt.Errorf("no SRV records found")
	}

	// Use the first SRV record (should sort by priority/weight)
	srv := srvs[0]

	return &models.ServerInfo{
		Host: srv.Target,
		Port: int(srv.Port),
	}, nil
}

// cacheServerInfo caches server information
func (r *ServerResolver) cacheServerInfo(domain string, serverInfo *models.ServerInfo) {
	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()

	r.cache[domain] = &cacheEntry{
		serverInfo: serverInfo,
		expiresAt:  time.Now().Add(1 * time.Hour),
	}

	logger.Debug("Cached server info",
		zap.String("domain", domain),
		zap.String("host", serverInfo.Host),
		zap.Int("port", serverInfo.Port))
}

// ClearCache clears the resolver cache
func (r *ServerResolver) ClearCache() {
	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()
	r.cache = make(map[string]*cacheEntry)
	logger.Info("Server resolver cache cleared")
}
