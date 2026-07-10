package tkPresentationMiddlewareHoneypot

import (
	"math/rand"
	"time"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

type HoneypotPathSelector struct {
	settings  tkDto.HoneypotSettings
	pathPool  HoneypotPathPool
	randomGen *rand.Rand
}

func NewHoneypotPathSelector(
	settings tkDto.HoneypotSettings,
	pathPool HoneypotPathPool,
) *HoneypotPathSelector {
	seedSource := time.Now().UnixNano()
	if settings.RandomSeed != 0 {
		seedSource = settings.RandomSeed
	}

	return &HoneypotPathSelector{
		settings:  settings,
		pathPool:  pathPool,
		randomGen: rand.New(rand.NewSource(seedSource)),
	}
}

func (selector *HoneypotPathSelector) Select() map[tkValueObject.UrlPath]tkValueObject.HoneypotPathClass {
	classDistribution := selector.resolveClassDistribution()

	orderedClasses := []tkValueObject.HoneypotPathClass{
		tkValueObject.HoneypotPathClassStaticVulnerability,
		tkValueObject.HoneypotPathClassBandwidthExhaust,
		tkValueObject.HoneypotPathClassAiTrap,
	}

	activePaths := make(
		map[tkValueObject.UrlPath]tkValueObject.HoneypotPathClass,
	)

	for _, class := range orderedClasses {
		count := classDistribution[class]
		classPaths := selector.pathPool.PathsByClass(class)
		if len(classPaths) == 0 {
			continue
		}

		shuffledPaths := make([]tkValueObject.UrlPath, len(classPaths))
		copy(shuffledPaths, classPaths)
		selector.randomGen.Shuffle(
			len(shuffledPaths),
			func(leftIdx, rightIdx int) {
				shuffledPaths[leftIdx], shuffledPaths[rightIdx] =
					shuffledPaths[rightIdx], shuffledPaths[leftIdx]
			},
		)

		selectedCount := count
		if selectedCount > len(shuffledPaths) {
			selectedCount = len(shuffledPaths)
		}

		for pathIdx := 0; pathIdx < selectedCount; pathIdx++ {
			activePaths[shuffledPaths[pathIdx]] = class
		}
	}

	return activePaths
}

func (selector *HoneypotPathSelector) resolveClassDistribution() map[tkValueObject.HoneypotPathClass]int {
	activePathCount := selector.settings.ActivePathCount.Int()
	totalPoolSize := selector.pathPool.Size()

	allClasses := []tkValueObject.HoneypotPathClass{
		tkValueObject.HoneypotPathClassStaticVulnerability,
		tkValueObject.HoneypotPathClassBandwidthExhaust,
		tkValueObject.HoneypotPathClassAiTrap,
	}

	distribution := make(map[tkValueObject.HoneypotPathClass]int)
	allocatedTotal := 0

	for _, class := range allClasses {
		classSize := len(selector.pathPool.PathsByClass(class))
		allocated := classSize * activePathCount / totalPoolSize
		distribution[class] = allocated
		allocatedTotal += allocated
	}

	remainder := activePathCount - allocatedTotal
	if remainder > 0 {
		largestClass := allClasses[0]
		largestSize := len(selector.pathPool.PathsByClass(largestClass))
		for _, class := range allClasses[1:] {
			classSize := len(selector.pathPool.PathsByClass(class))
			if classSize > largestSize {
				largestClass = class
				largestSize = classSize
			}
		}
		distribution[largestClass] += remainder
	}

	return distribution
}
