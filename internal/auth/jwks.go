package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"sync"
	"time"
)

type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwksJSON struct {
	Keys []jwkKey `json:"keys"`
}

// JWKS caches RSA public keys by kid.
type JWKS struct {
	url    string
	mu     sync.RWMutex
	keys   map[string]*rsa.PublicKey
	ticker *time.Ticker
	quit   chan struct{}
}

// NewJWKS creates a JWKS instance and loads keys immediately. It also starts
// a background refresh every refreshInterval. Pass 0 to use default 15m.
func NewJWKS(url string, refreshInterval time.Duration) (*JWKS, error) {
	if refreshInterval <= 0 {
		refreshInterval = 15 * time.Minute
	}
	j := &JWKS{
		url:    url,
		keys:   map[string]*rsa.PublicKey{},
		ticker: time.NewTicker(refreshInterval),
		quit:   make(chan struct{}),
	}
	if err := j.refresh(); err != nil {
		return nil, err
	}
	go j.loop()
	return j, nil
}

func (j *JWKS) loop() {
	for {
		select {
		case <-j.ticker.C:
			_ = j.refresh()
		case <-j.quit:
			return
		}
	}
}

// Close stops background refresh.
func (j *JWKS) Close() {
	close(j.quit)
	j.ticker.Stop()
}

func (j *JWKS) refresh() error {
	resp, err := http.Get(j.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var raw jwksJSON
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return err
	}

	newKeys := make(map[string]*rsa.PublicKey)
	for _, k := range raw.Keys {
		if k.Kty != "RSA" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			return err
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			return err
		}
		n := new(big.Int).SetBytes(nBytes)
		e := bytesToInt(eBytes)
		pub := &rsa.PublicKey{
			N: n,
			E: e,
		}
		newKeys[k.Kid] = pub
	}

	j.mu.Lock()
	defer j.mu.Unlock()
	j.keys = newKeys
	return nil
}

func (j *JWKS) Get(kid string) (*rsa.PublicKey, error) {
	j.mu.RLock()
	p := j.keys[kid]
	j.mu.RUnlock()
	if p != nil {
		return p, nil
	}
	// try refresh once
	if err := j.refresh(); err != nil {
		return nil, err
	}
	j.mu.RLock()
	defer j.mu.RUnlock()
	p = j.keys[kid]
	if p == nil {
		return nil, errors.New("jwks: key not found")
	}
	return p, nil
}

func bytesToInt(b []byte) int {
	res := 0
	for _, v := range b {
		res = (res << 8) + int(v)
	}
	return res
}
