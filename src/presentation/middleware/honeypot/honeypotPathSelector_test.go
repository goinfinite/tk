package tkPresentationMiddlewareHoneypot

import (
	"sort"
	"testing"

	tkDto "github.com/goinfinite/tk/src/domain/dto"
	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestHoneypotPathSelector(t *testing.T) {
	pathPool := NewHoneypotPathPool()

	t.Run("SelectsExactlyActivePathCount", func(t *testing.T) {
		activePathCount, _ := tkValueObject.NewHoneypotActivePathCount(
			30, pathPool.Size(),
		)
		settings := tkDto.HoneypotSettings{
			ActivePathCount: activePathCount,
			RandomSeed:      42,
		}

		selector := NewHoneypotPathSelector(settings, pathPool)
		activePaths := selector.Select()

		if len(activePaths) != 30 {
			t.Errorf(
				"SelectedPathCountMismatch: Expected=30, Actual=%d",
				len(activePaths),
			)
		}
	})

	t.Run("DeterministicWithSeed", func(t *testing.T) {
		activePathCount, _ := tkValueObject.NewHoneypotActivePathCount(
			30, pathPool.Size(),
		)
		settings := tkDto.HoneypotSettings{
			ActivePathCount: activePathCount,
			RandomSeed:      42,
		}

		selectorA := NewHoneypotPathSelector(settings, pathPool)
		pathsA := selectorA.Select()

		selectorB := NewHoneypotPathSelector(settings, pathPool)
		pathsB := selectorB.Select()

		if len(pathsA) != len(pathsB) {
			t.Errorf(
				"DeterministicCountMismatch: A=%d, B=%d",
				len(pathsA), len(pathsB),
			)
		}

		sortedPathsA := make([]string, 0, len(pathsA))
		for path := range pathsA {
			sortedPathsA = append(sortedPathsA, path.String())
		}
		sort.Strings(sortedPathsA)

		sortedPathsB := make([]string, 0, len(pathsB))
		for path := range pathsB {
			sortedPathsB = append(sortedPathsB, path.String())
		}
		sort.Strings(sortedPathsB)

		for pathIdx, pathA := range sortedPathsA {
			if pathA != sortedPathsB[pathIdx] {
				t.Errorf(
					"DeterministicPathMismatch: A='%s', B='%s'",
					pathA, sortedPathsB[pathIdx],
				)
			}
		}
	})

	t.Run("ActivePathCountClampedToFloor", func(t *testing.T) {
		activePathCount, _ := tkValueObject.NewHoneypotActivePathCount(
			0, pathPool.Size(),
		)
		settings := tkDto.HoneypotSettings{
			ActivePathCount: activePathCount,
			RandomSeed:      42,
		}

		selector := NewHoneypotPathSelector(settings, pathPool)
		activePaths := selector.Select()

		if len(activePaths) != 30 {
			t.Errorf(
				"FloorClampedCountMismatch: Expected=30, Actual=%d",
				len(activePaths),
			)
		}
	})
}
