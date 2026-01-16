package config

import "encoding/json"

type RouterConfig struct {
	Flow FlowConfig `json:"flow"`
}

type FlowConfig struct {
	Action  string
	Params  any
	Subflow []FlowConfig
}

type FlowConfigJSON struct {
	Action  string          `json:"action"`
	Params  json.RawMessage `json:"params"`
	Subflow []FlowConfig    `json:"subflow"`
}

func (c *FlowConfig) UnmarshalJSON(data []byte) error {
	var cfgJSON FlowConfigJSON
	err := json.Unmarshal(data, &cfgJSON)
	if err != nil {
		return err
	}
	c.Action = cfgJSON.Action
	c.Subflow = cfgJSON.Subflow

	switch cfgJSON.Action {
	case "match":
		var p MatchParams
		err := json.Unmarshal(cfgJSON.Params, &p)
		if err != nil {
			return err
		}
		c.Params = p
	case "dispatch":
		var p DispatchParams
		err := json.Unmarshal(cfgJSON.Params, &p)
		if err != nil {
			return err
		}
		c.Params = p
	default:
	}
	return nil
}

type MatchParams struct {
	Matchers []MatcherConfig `json:"matchers"`
}

type DispatchParams struct {
	Analyzer string `json:"analyzer"`
}
