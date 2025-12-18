package dto

type Client struct {
	TotalSent int64
	Resources map[string]Resource
}
