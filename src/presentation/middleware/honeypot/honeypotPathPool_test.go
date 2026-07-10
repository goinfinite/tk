package tkPresentationMiddlewareHoneypot

import (
	"testing"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestHoneypotPathPool(t *testing.T) {
	pool := NewHoneypotPathPool()

	t.Run("PathsInAllThreeClasses", func(t *testing.T) {
		staticPaths := pool.PathsByClass(
			tkValueObject.HoneypotPathClassStaticVulnerability,
		)
		if len(staticPaths) == 0 {
			t.Errorf("StaticVulnerabilityPathsEmpty: ExpectedNonEmpty")
		}

		bandwidthPaths := pool.PathsByClass(
			tkValueObject.HoneypotPathClassBandwidthExhaust,
		)
		if len(bandwidthPaths) == 0 {
			t.Errorf("BandwidthExhaustPathsEmpty: ExpectedNonEmpty")
		}

		aiTrapPaths := pool.PathsByClass(
			tkValueObject.HoneypotPathClassAiTrap,
		)
		if len(aiTrapPaths) == 0 {
			t.Errorf("AiTrapPathsEmpty: ExpectedNonEmpty")
		}
	})

	t.Run("NoOverlapBetweenClasses", func(t *testing.T) {
		staticPaths := pool.PathsByClass(
			tkValueObject.HoneypotPathClassStaticVulnerability,
		)
		bandwidthPaths := pool.PathsByClass(
			tkValueObject.HoneypotPathClassBandwidthExhaust,
		)
		aiTrapPaths := pool.PathsByClass(
			tkValueObject.HoneypotPathClassAiTrap,
		)

		pathSet := make(map[string]bool)
		for _, path := range staticPaths {
			pathSet[path.String()] = true
		}
		for _, path := range bandwidthPaths {
			if pathSet[path.String()] {
				t.Errorf(
					"PathOverlapInBandwidthExhaust: '%s'",
					path.String(),
				)
			}
		}
		for _, path := range aiTrapPaths {
			if pathSet[path.String()] {
				t.Errorf("PathOverlapInAiTrap: '%s'", path.String())
			}
		}
	})

	t.Run("SizeEqualsSumOfAllClasses", func(t *testing.T) {
		staticCount := len(pool.PathsByClass(
			tkValueObject.HoneypotPathClassStaticVulnerability,
		))
		bandwidthCount := len(pool.PathsByClass(
			tkValueObject.HoneypotPathClassBandwidthExhaust,
		))
		aiTrapCount := len(pool.PathsByClass(
			tkValueObject.HoneypotPathClassAiTrap,
		))
		expectedSize := staticCount + bandwidthCount + aiTrapCount

		if pool.Size() != expectedSize {
			t.Errorf(
				"PoolSizeMismatch: Expected=%d, Actual=%d",
				expectedSize, pool.Size(),
			)
		}
	})
}
