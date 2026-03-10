package rag

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cy77cc/OpsPilot/internal/config"
	milvusclient "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type MilvusClient struct {
	cfg config.Milvus

	mu     sync.RWMutex
	client milvusclient.Client
}

const defaultMilvusOpTimeout = 10 * time.Second

func NewMilvusClient(cfg config.Milvus) *MilvusClient {
	return &MilvusClient{cfg: cfg}
}

func (m *MilvusClient) operationTimeout() time.Duration {
	if m == nil || m.cfg.Timeout <= 0 {
		return defaultMilvusOpTimeout
	}
	return m.cfg.Timeout
}

func (m *MilvusClient) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, m.operationTimeout())
}

func (m *MilvusClient) Connect(ctx context.Context) error {
	if m == nil {
		return fmt.Errorf("milvus client is nil")
	}
	opCtx, cancel := m.withTimeout(ctx)
	defer cancel()
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.client != nil {
		return nil
	}
	client, err := milvusclient.NewClient(opCtx, milvusclient.Config{
		Address:       fmt.Sprintf("%s:%s", m.cfg.Host, m.cfg.Port),
		Username:      m.cfg.Username,
		Password:      m.cfg.Password,
		DBName:        m.cfg.Database,
		EnableTLSAuth: m.cfg.UseTLS,
	})
	if err != nil {
		return err
	}
	m.client = client
	return nil
}

func (m *MilvusClient) Close() error {
	if m == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.client == nil {
		return nil
	}
	err := m.client.Close()
	m.client = nil
	return err
}

func (m *MilvusClient) CheckHealth(ctx context.Context) error {
	if m == nil {
		return fmt.Errorf("milvus client is nil")
	}
	opCtx, cancel := m.withTimeout(ctx)
	defer cancel()
	m.mu.RLock()
	client := m.client
	m.mu.RUnlock()
	if client == nil {
		return fmt.Errorf("milvus client not connected")
	}
	state, err := client.CheckHealth(opCtx)
	if err != nil {
		return err
	}
	if state == nil || !state.IsHealthy {
		return fmt.Errorf("milvus unhealthy")
	}
	return nil
}

func (m *MilvusClient) Reconnect(ctx context.Context) error {
	if m == nil {
		return fmt.Errorf("milvus client is nil")
	}
	_ = m.Close()
	return m.Connect(ctx)
}

func (m *MilvusClient) StartHealthMonitor(ctx context.Context) {
	if m == nil {
		return
	}
	period := m.cfg.HealthCheckInterval
	if period <= 0 {
		period = 15 * time.Second
	}
	go func() {
		ticker := time.NewTicker(period)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				checkCtx, cancel := context.WithTimeout(ctx, m.operationTimeout())
				err := m.CheckHealth(checkCtx)
				cancel()
				if err != nil {
					reconnectCtx, rcancel := context.WithTimeout(ctx, m.operationTimeout())
					_ = m.Reconnect(reconnectCtx)
					rcancel()
				}
			}
		}
	}()
}

func (m *MilvusClient) InsertRows(ctx context.Context, collection string, rows []interface{}) error {
	if len(rows) == 0 {
		return nil
	}
	if err := m.Connect(ctx); err != nil {
		return err
	}
	m.mu.RLock()
	client := m.client
	m.mu.RUnlock()
	if client == nil {
		return fmt.Errorf("milvus client not connected")
	}
	opCtx, cancel := m.withTimeout(ctx)
	defer cancel()
	_, err := client.InsertRows(opCtx, collection, "", rows)
	return err
}

func (m *MilvusClient) FlushCollection(ctx context.Context, collection string) error {
	if err := m.Connect(ctx); err != nil {
		return err
	}
	m.mu.RLock()
	client := m.client
	m.mu.RUnlock()
	if client == nil {
		return fmt.Errorf("milvus client not connected")
	}
	opCtx, cancel := m.withTimeout(ctx)
	defer cancel()
	return client.Flush(opCtx, collection, false)
}

func (m *MilvusClient) LoadCollection(ctx context.Context, collection string) error {
	if err := m.Connect(ctx); err != nil {
		return err
	}
	m.mu.RLock()
	client := m.client
	m.mu.RUnlock()
	if client == nil {
		return fmt.Errorf("milvus client not connected")
	}
	opCtx, cancel := m.withTimeout(ctx)
	defer cancel()
	return client.LoadCollection(opCtx, collection, false)
}

func (m *MilvusClient) SearchVectors(ctx context.Context, collection, vectorField string, vectors []entity.Vector, topK int) ([]milvusclient.SearchResult, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}
	m.mu.RLock()
	client := m.client
	m.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("milvus client not connected")
	}
	opCtx, cancel := m.withTimeout(ctx)
	defer cancel()
	searchParam, err := entity.NewIndexFlatSearchParam()
	if err != nil {
		return nil, err
	}
	return client.Search(opCtx, collection, nil, "", nil, vectors, vectorField, entity.COSINE, topK, searchParam)
}

func getenv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func getenvInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func getenvBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if raw == "" {
		return fallback
	}
	switch raw {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return d
}
