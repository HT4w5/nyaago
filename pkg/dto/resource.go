package dto

type Resource struct {
	URL       string
	Size      int64
	TotalSent int64
	SendRatio float64 // Average times of target resource sent per unit time
}
