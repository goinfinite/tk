package tkRepository

import tkDto "github.com/goinfinite/tk/src/domain/dto"

type HoneypotQueryRepo interface {
	ReadBanDecision(tkDto.ReadHoneypotBanDecisionRequest) (tkDto.ReadHoneypotBanDecisionResponse, error)
	ReadStatsReport(tkDto.ReadHoneypotStatsReportRequest) (tkDto.ReadHoneypotStatsReportResponse, error)
}
