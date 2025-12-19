package pool

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/HT4w5/nyaago/pkg/db"
	"github.com/HT4w5/nyaago/pkg/dto"
	"go4.org/netipx"
)

func (p *Pool) GetRules() ([]dto.Rule, error) {
	rules, err := p.adapter.ListRules()
	if err != nil {
		return []dto.Rule{}, fmt.Errorf("failed to list rules: %w", err)
	}

	var result []dto.Rule

	for _, v := range rules {
		result = append(result, dto.Rule{
			Prefix:    v.Prefix,
			Addr:      v.Addr,
			URL:       v.URL,
			ExpiresOn: v.ExpiresOn,
		})
	}

	return result, nil
}

func (p *Pool) GetRuleSet() (*netipx.IPSet, error) {
	rules, err := p.adapter.ListRules()
	if err != nil {
		return nil, fmt.Errorf("failed to list rules: %w", err)
	}

	var b netipx.IPSetBuilder
	for _, v := range rules {
		b.AddPrefix(v.Prefix)
	}

	return b.IPSet()
}

func (p *Pool) FlushExpired() error {
	var err error
	err = p.adapter.FlushExpiredClients()
	if err != nil {
		return fmt.Errorf("failed to flush expired clients: %w", err)
	}
	err = p.adapter.FlushExpiredRequests()
	if err != nil {
		return fmt.Errorf("failed to flush expired requests: %w", err)
	}
	err = p.adapter.FlushExpiredResources()
	if err != nil {
		return fmt.Errorf("failed to flush expired resources: %w", err)
	}
	err = p.adapter.FlushExpiredRules()
	if err != nil {
		return fmt.Errorf("failed to flush expired rules: %w", err)
	}

	return nil
}

func (p *Pool) BuildRules() error {
	// Filter for malicious requests
	requests, err := p.adapter.FilterRequests(p.cfg.Pool.SendRatioThreshold, p.cfg.Pool.RequestConfig.MaturationThreshold)
	if err != nil {
		return fmt.Errorf("failed to filter for requests: %w", err)
	}

	// Generate and store rules
	currentTime := time.Now()
	for _, v := range requests {
		var prefixLength int
		if v.Addr.Is4() {
			prefixLength = p.cfg.Pool.BanPrefixLength.IPv4
		} else if v.Addr.Is6() {
			prefixLength = p.cfg.Pool.BanPrefixLength.IPv6
		} else {
			// Drop invalid
			continue
		}

		r := db.Rule{
			Prefix:    netip.PrefixFrom(v.Addr, prefixLength),
			Addr:      v.Addr,
			URL:       v.URL,
			ExpiresOn: currentTime.Add(p.cfg.Pool.RuleConfig.TTL.Duration),
		}

		err = p.adapter.PutRule(r)
		if err != nil {
			return fmt.Errorf("failed to put rule: %w", err)
		}
	}

	return nil
}
