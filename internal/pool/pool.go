package pool

import (
	"fmt"
	"time"

	"github.com/HT4w5/nyaago/internal/config"
	"github.com/HT4w5/nyaago/pkg/db"
	"github.com/HT4w5/nyaago/pkg/dto"
)

var pool *Pool

type Pool struct {
	cfg     *config.Config
	adapter db.DBAdapter
}

func GetPool(cfg *config.Config) (*Pool, error) {
	if pool != nil {
		return pool, nil
	}

	// Make DBAdapter
	db, err := db.MakeDBAdapter(cfg.DB.Type, cfg.DB.Access)
	if err != nil {
		return nil, fmt.Errorf("failed to get DBAdapter: %w", err)
	}

	p := &Pool{
		cfg:     cfg,
		adapter: db,
	}

	pool = p
	return p, nil
}

func (p *Pool) ProcessRequest(req dto.Request) error {
	// -- Apply filters --
	// Apply include first, then exclude
	include := false
	for _, v := range p.cfg.Analyzer.Include {
		if v.Match(req) {
			include = true
			break
		}
	}
	if !include {
		return nil
	}

	for _, v := range p.cfg.Analyzer.Exclude {
		if v.Match(req) {
			return nil
		}
	}

	currentTime := time.Now()

	// -- Process client --
	client, err := p.adapter.GetClient(req.Client)
	if err != nil {
		return fmt.Errorf("failed to get client %s: %w", req.Client, err)
	}
	if client.Addr.IsValid() {
		// Update client attributes
		client.TotalSent += req.BodySent
		client.ExpiresOn = currentTime.Add(p.cfg.Pool.ClientConfig.TTL.Duration)
	} else {
		// Create new
		client = db.Client{
			Addr:      req.Client,
			TotalSent: req.BodySent,
			CreatedOn: currentTime,
			ExpiresOn: currentTime.Add(p.cfg.Pool.ClientConfig.TTL.Duration),
		}
	}

	err = p.adapter.PutClient(client)
	if err != nil {
		return fmt.Errorf("failed to put client %s: %w", client.Addr, err)
	}

	// -- Process resource --
	resource, err := p.adapter.GetResource(req.URL)
	if err != nil {
		return fmt.Errorf("failed to get resource %s: %w", req.URL, err)
	}
	// Update resource attributes
	resource.Size = max(resource.Size, req.BodySent)
	resource.ExpiresOn = currentTime.Add(p.cfg.Pool.ResourceConfig.TTL.Duration)
	resource.URL = req.URL

	err = p.adapter.PutResource(resource)
	if err != nil {
		return fmt.Errorf("failed to put resource %s: %w", client.Addr, err)
	}

	// -- Process request --
	request, err := p.adapter.GetRequest(req.Client, req.URL)
	if request.Addr.IsValid() {
		// Update request attributes
		request.ExpiresOn = currentTime.Add(p.cfg.Pool.ResourceConfig.TTL.Duration)
		request.TotalSent += req.BodySent
		request.SendRatio = float64(request.TotalSent*24) / float64(resource.Size) / currentTime.Sub(request.CreatedOn).Hours() // Sent times per day
		request.Occurrence++
	} else {
		// Create new
		request = db.Request{
			Addr:       req.Client,
			URL:        req.URL,
			TotalSent:  req.BodySent,
			SendRatio:  1.,
			Occurrence: 1,
			CreatedOn:  currentTime,
			ExpiresOn:  currentTime.Add(p.cfg.Pool.RequestConfig.TTL.Duration),
		}
	}

	err = p.adapter.PutRequest(request)
	if err != nil {
		return fmt.Errorf("failed to put request %s: %w", client.Addr, err)
	}

	return nil
}
