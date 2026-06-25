package tkInfraHoneypot

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
	tkInfraDb "github.com/goinfinite/tk/src/infra/db"
)

type HoneypotQueryRepo struct {
	transientDbSvc *tkInfraDb.TransientDatabaseService
}

func NewHoneypotQueryRepo(
	transientDbSvc *tkInfraDb.TransientDatabaseService,
) *HoneypotQueryRepo {
	return &HoneypotQueryRepo{transientDbSvc: transientDbSvc}
}

func buildTopOffenders(
	ipHitCounts map[string]int,
	limit int,
) []tkDto.HoneypotStatsOffender {
	offenders := make(
		[]tkDto.HoneypotStatsOffender, 0, len(ipHitCounts),
	)
	for ipAddr, hitCount := range ipHitCounts {
		offenders = append(offenders, tkDto.HoneypotStatsOffender{
			HitCount:  hitCount,
			IpAddress: ipAddr,
		})
	}
	sort.Slice(offenders, func(a, b int) bool {
		return offenders[a].HitCount > offenders[b].HitCount
	})
	if len(offenders) > limit {
		offenders = offenders[:limit]
	}
	return offenders
}

func buildTopEndpoints(
	endpointCounts map[string]int,
	limit int,
) []tkDto.HoneypotStatsEndpoint {
	endpoints := make(
		[]tkDto.HoneypotStatsEndpoint, 0, len(endpointCounts),
	)
	for endpointPath, hitCount := range endpointCounts {
		endpoints = append(
			endpoints, tkDto.HoneypotStatsEndpoint{
				HitCount: hitCount,
				Path:     endpointPath,
			},
		)
	}
	sort.Slice(endpoints, func(a, b int) bool {
		return endpoints[a].HitCount > endpoints[b].HitCount
	})
	if len(endpoints) > limit {
		endpoints = endpoints[:limit]
	}
	return endpoints
}

func (repo *HoneypotQueryRepo) ReadHitRecord(
	requesterIp tkValueObject.IpAddress,
) (tkDto.HoneypotHitData, error) {
	var hitData tkDto.HoneypotHitData
	if repo.transientDbSvc == nil {
		return hitData, errors.New("TransientDbUnavailable")
	}

	hitKey := honeypotHitKeyPrefix + requesterIp.String()
	rawValue, readErr := repo.transientDbSvc.Read(hitKey)
	if readErr != nil {
		return hitData, readErr
	}

	parseErr := json.Unmarshal([]byte(rawValue), &hitData)
	return hitData, parseErr
}

func (repo *HoneypotQueryRepo) Count() int64 {
	if repo.transientDbSvc == nil {
		return 0
	}
	return repo.transientDbSvc.Count()
}

func (repo *HoneypotQueryRepo) ReadReport(
	banDuration time.Duration,
	aggressivenessMode tkValueObject.HoneypotAggressivenessMode,
) (tkDto.HoneypotStatsReport, error) {
	var report tkDto.HoneypotStatsReport
	if repo.transientDbSvc == nil {
		return report, nil
	}

	entries, readAllErr := repo.transientDbSvc.ReadAll()
	if readAllErr != nil {
		return report, readAllErr
	}

	ipHitCounts := make(map[string]int)
	endpointCounts := make(map[string]int)
	bannedIpCount := 0

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Key, honeypotHitKeyPrefix) {
			continue
		}

		var hitData tkDto.HoneypotHitData
		parseErr := json.Unmarshal(
			[]byte(entry.Value), &hitData,
		)
		if parseErr != nil {
			continue
		}

		firstHitAt, timeParseErr := time.Parse(
			time.RFC3339, hitData.FirstHitAt,
		)
		if timeParseErr != nil {
			continue
		}

		banWindowStart := time.Now().Add(-banDuration)
		if firstHitAt.Before(banWindowStart) {
			continue
		}

		ipAddr := strings.TrimPrefix(
			entry.Key, honeypotHitKeyPrefix,
		)
		resolvedTier := aggressivenessMode.ResolveTier(
			hitData.Count,
		)
		if resolvedTier >= 2 {
			bannedIpCount++
		}

		ipHitCounts[ipAddr] = hitData.Count
		for endpointPath, count := range hitData.Endpoints {
			endpointCounts[endpointPath] += count
		}
	}

	report.BannedIpCount = bannedIpCount
	report.TopEndpoints = buildTopEndpoints(
		endpointCounts, 10,
	)
	report.TopOffenders = buildTopOffenders(
		ipHitCounts, 10,
	)
	return report, nil
}
