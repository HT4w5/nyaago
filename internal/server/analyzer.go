package server

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/HT4w5/nyaago/internal/logging"
	"github.com/HT4w5/nyaago/pkg/db"
	"github.com/HT4w5/nyaago/pkg/dto"
	"go4.org/netipx"
)

func (s *Server) processRequest(req dto.Request) {
	s.logger.Debug("processing request", "request", req)
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
		return
	}

	for _, v := range s.cfg.Analyzer.Exclude {
		if v.Match(req) {
			return
		}
	}

	currentTime := time.Now()

	// -- Process client --
	client, err := s.db.GetClient(req.Client)
	if err != nil {
		s.logger.Error("failed to get client", logging.SlogKeyError, err, "addr", req.Client)
		return
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
		s.logger.Error("failed to start db transaction", logging.SlogKeyError, err)
		return
	}

	defer func() {
		err := tx.Commit()
		if err != nil {
			s.logger.Error("failed to commit db transaction", logging.SlogKeyError, err)
		}
	}()

	err = tx.PutClient(client)
	if err != nil {
		s.logger.Error("failed to put client", logging.SlogKeyError, err, "client", client)
		return
	}

	// -- Process resource --
	resource, err := s.db.GetResource(req.URL)
	if err != nil {
		s.logger.Error("failed to get resource", logging.SlogKeyError, err, "url", req.URL)
		return
	}
	// Update resource attributes
	resource.Size = max(resource.Size, req.BodySent)
	resource.ExpiresOn = currentTime.Add(s.cfg.Analyzer.TTL.Duration)
	resource.URL = req.URL

	err = tx.PutResource(resource)
	if err != nil {
		s.logger.Error("failed to put resource", logging.SlogKeyError, err, "resource", resource)
		return
	}

	// -- Process request --
	request, err := s.db.GetRequest(req.Client, req.URL)
	if request.Addr.IsValid() {
		// Update request attributes
		request.ExpiresOn = currentTime.Add(s.cfg.Analyzer.TTL.Duration)
		request.TotalSent += req.BodySent
		// request.SendRatio = float64(request.TotalSent*24) / float64(resource.Size) / currentTime.Sub(request.CreatedOn).Hours() // Sent times per day
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
		s.logger.Error("failed to put request", logging.SlogKeyError, err, "request", request)
		return
	}
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

func (s *Server) flushExpired() {
	s.logger.Info("flushing expired pool objects")
	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Error("failed to start db transaction", logging.SlogKeyError, err)
		return
	}

	defer func() {
		err := tx.Commit()
		if err != nil {
			s.logger.Error("failed to commit db transaction", logging.SlogKeyError, err)
		}
	}()

	err = tx.FlushExpiredClients()
	if err != nil {
		s.logger.Error("failed to flush expired clients", logging.SlogKeyError, err)
	}
	err = tx.FlushExpiredRequests()
	if err != nil {
		s.logger.Error("failed to flush expired requests", logging.SlogKeyError, err)
	}
	err = tx.FlushExpiredResources()
	if err != nil {
		s.logger.Error("failed to flush expired resources", logging.SlogKeyError, err)
	}
	err = tx.FlushExpiredRules()
	if err != nil {
		s.logger.Error("failed to flush expired rules", logging.SlogKeyError, err)
	}
}

func (s *Server) computeRules() {
	s.logger.Info("computing new rules")
	// Filter for mature requests (older than one update interval)
	currentTime := time.Now()
	requests, err := s.db.ListRequests(currentTime.Add(-s.cfg.Analyzer.UpdateInterval.Duration))
	if err != nil {
		s.logger.Error("failed to list requests", logging.SlogKeyError, err)
		return
	}

	// Recompute SendRatio for each request
	cache := makeSizeCache(s.db)

	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Error("failed to start db transaction", logging.SlogKeyError, err)
		return
	}

	defer func() {
		err := tx.Commit()
		if err != nil {
			s.logger.Error("failed to commit db transaction", logging.SlogKeyError, err)
		}
	}()

	candidates := make([]db.Request, 0, len(requests))

	for _, v := range requests {
		size, err := cache.GetSize(v.URL)
		if err != nil {
			s.logger.Error("failed to get resource size", logging.SlogKeyError, err)
			continue
		}
		if size <= 0 {
			s.logger.Warn("resource not found")
			continue
		}

		v.SendRatio = float64(v.TotalSent*24) / float64(size) / currentTime.Sub(v.CreatedOn).Hours() // Sent times per day

		if v.SendRatio >= s.cfg.Analyzer.SendRatioThreshold {
			candidates = append(candidates, v)
		}

		err = tx.PutRequest(v)
		if err != nil {
			s.logger.Error("failed to put request", logging.SlogKeyError, err)
		}
	}

	// Compute and store rules
	for _, v := range candidates {
		var prefixLength int
		if v.Addr.Is4() {
			prefixLength = s.cfg.Analyzer.BanPrefixLength.IPv4
		} else if v.Addr.Is6() {
			prefixLength = s.cfg.Analyzer.BanPrefixLength.IPv6
		} else {
			// Drop invalid
			s.logger.Warn("dropped request with invalid addr", "addr", v.Addr)
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
			s.logger.Error("failed to put rule", logging.SlogKeyError, err)
		}
	}
}
