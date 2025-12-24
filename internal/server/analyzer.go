package server

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/HT4w5/nyaago/pkg/db"
	"github.com/HT4w5/nyaago/pkg/dto"
	"go4.org/netipx"
)

func (s *Server) processRequest(req dto.Request) error {
	// -- Apply filters --
	// Apply include first, then exclude
	include := false
	for _, v := range s.cfg.Analyzer.Include {
		if v.Match(req) {
			include = true
			break
		}
	}
	if !include {
		return nil
	}

	for _, v := range s.cfg.Analyzer.Exclude {
		if v.Match(req) {
			return nil
		}
	}

	currentTime := time.Now()

	// -- Process client --
	client, err := s.db.GetClient(req.Client)
	if err != nil {
		return fmt.Errorf("failed to get client %s: %w", req.Client, err)
	}
	if client.Addr.IsValid() {
		// Update client attributes
		client.TotalSent += req.BodySent
		client.ExpiresOn = currentTime.Add(s.cfg.Analyzer.TTL.Duration)
	} else {
		// Create new
		client = db.Client{
			Addr:      req.Client,
			TotalSent: req.BodySent,
			CreatedOn: req.Time,
			ExpiresOn: currentTime.Add(s.cfg.Analyzer.TTL.Duration),
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start db transaction: %w", err)
	}

	err = tx.PutClient(client)
	if err != nil {
		return fmt.Errorf("failed to put client %s: %w", client.Addr, err)
	}

	// -- Process resource --
	resource, err := s.db.GetResource(req.URL)
	if err != nil {
		return fmt.Errorf("failed to get resource %s: %w", req.URL, err)
	}
	// Update resource attributes
	resource.Size = max(resource.Size, req.BodySent)
	resource.ExpiresOn = currentTime.Add(s.cfg.Analyzer.TTL.Duration)
	resource.URL = req.URL

	err = tx.PutResource(resource)
	if err != nil {
		return fmt.Errorf("failed to put resource %s: %w", client.Addr, err)
	}

	// -- Process request --
	request, err := s.db.GetRequest(req.Client, req.URL)
	if request.Addr.IsValid() {
		// Update request attributes
		request.ExpiresOn = currentTime.Add(s.cfg.Analyzer.TTL.Duration)
		request.TotalSent += req.BodySent
		request.SendRatio = float64(request.TotalSent*24) / float64(resource.Size) / currentTime.Sub(request.CreatedOn).Hours() // Sent times per day
	} else {
		// Create new
		request = db.Request{
			Addr:      req.Client,
			URL:       req.URL,
			TotalSent: req.BodySent,
			SendRatio: 0.,
			CreatedOn: req.Time,
			ExpiresOn: currentTime.Add(s.cfg.Analyzer.TTL.Duration),
		}
	}

	err = tx.PutRequest(request)
	if err != nil {
		return fmt.Errorf("failed to put request %s: %w", client.Addr, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit db transaction: %w", err)
	}

	return nil
}

func (s *Server) getRuleSet() (*netipx.IPSet, error) {
	rules, err := s.db.ListRules()
	if err != nil {
		return nil, fmt.Errorf("failed to list rules: %w", err)
	}

	var b netipx.IPSetBuilder
	for _, v := range rules {
		b.AddPrefix(v.Prefix)
	}

	return b.IPSet()
}

func (s *Server) flushExpired() error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start db transaction: %w", err)
	}
	err = tx.FlushExpiredClients()
	if err != nil {
		return fmt.Errorf("failed to flush expired clients: %w", err)
	}
	err = tx.FlushExpiredRequests()
	if err != nil {
		return fmt.Errorf("failed to flush expired requests: %w", err)
	}
	err = tx.FlushExpiredResources()
	if err != nil {
		return fmt.Errorf("failed to flush expired resources: %w", err)
	}
	err = tx.FlushExpiredRules()
	if err != nil {
		return fmt.Errorf("failed to flush expired rules: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit db transaction: %w", err)
	}

	return nil
}

func (s *Server) buildRules() error {
	// Filter for malicious requests
	currentTime := time.Now()
	requests, err := s.db.FilterRequests(s.cfg.Analyzer.SendRatioThreshold, currentTime.Add(-s.cfg.Analyzer.UpdateInterval.Duration))
	if err != nil {
		return fmt.Errorf("failed to filter for requests: %w", err)
	}

	// Generate and store rules
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start db transaction: %w", err)
	}

	for _, v := range requests {
		var prefixLength int
		if v.Addr.Is4() {
			prefixLength = s.cfg.Analyzer.BanPrefixLength.IPv4
		} else if v.Addr.Is6() {
			prefixLength = s.cfg.Analyzer.BanPrefixLength.IPv6
		} else {
			// Drop invalid
			continue
		}

		r := dto.Rule{
			Prefix:    netip.PrefixFrom(v.Addr, prefixLength),
			Addr:      v.Addr,
			URL:       v.URL,
			ExpiresOn: currentTime.Add(s.cfg.Analyzer.TTL.Duration),
		}

		err = tx.PutRule(r)
		if err != nil {
			return fmt.Errorf("failed to put rule: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit db transaction: %w", err)
	}

	return nil
}
