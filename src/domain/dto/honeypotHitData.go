package tkDto

type HoneypotHitData struct {
	Count      int            `json:"count"`
	Endpoints  map[string]int `json:"endpoints"`
	FirstHitAt string         `json:"firstHitAt"`
}
