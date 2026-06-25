package tkDto

type HoneypotStatsEndpoint struct {
	HitCount int    `json:"hitCount"`
	Path     string `json:"path"`
}

type HoneypotStatsOffender struct {
	HitCount int    `json:"hitCount"`
	IpAddress string `json:"ipAddress"`
	Tier     int    `json:"tier"`
}

type HoneypotStatsReport struct {
	BannedIpCount int                       `json:"bannedIpCount"`
	TopEndpoints  []HoneypotStatsEndpoint   `json:"topEndpoints"`
	TopOffenders  []HoneypotStatsOffender   `json:"topOffenders"`
}
