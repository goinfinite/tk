package tkInfraDb

import (
	"sync"
	"testing"
)

func TestNewTransientDatabaseService(t *testing.T) {
	t.Run("ValidCreation", func(t *testing.T) {
		dbSvc, err := NewTransientDatabaseService()
		if err != nil {
			t.Errorf("UnexpectedError: %v", err)
			return
		}
		if dbSvc == nil {
			t.Errorf("ServiceIsNil")
			return
		}
		if dbSvc.Handler == nil {
			t.Errorf("HandlerIsNil")
		}
	})

	t.Run("ConcurrentCalls", func(t *testing.T) {
		const concurrentCount = 10

		var waitGroup sync.WaitGroup
		waitGroup.Add(concurrentCount)

		serviceHandlers := make([]*TransientDatabaseService, concurrentCount)
		serviceErrors := make([]error, concurrentCount)

		for callIndex := 0; callIndex < concurrentCount; callIndex++ {
			go func(index int) {
				defer waitGroup.Done()
				dbSvc, err := NewTransientDatabaseService()
				serviceHandlers[index] = dbSvc
				serviceErrors[index] = err
			}(callIndex)
		}

		waitGroup.Wait()

		for callIndex := 0; callIndex < concurrentCount; callIndex++ {
			if serviceErrors[callIndex] != nil {
				t.Errorf(
					"ConcurrentCall%dError: %v",
					callIndex,
					serviceErrors[callIndex],
				)
				return
			}
			if serviceHandlers[callIndex] == nil {
				t.Errorf("ConcurrentCall%dServiceIsNil", callIndex)
				return
			}
			if serviceHandlers[callIndex].Handler == nil {
				t.Errorf("ConcurrentCall%dHandlerIsNil", callIndex)
			}
		}

		for outerIndex := 0; outerIndex < concurrentCount; outerIndex++ {
			for innerIndex := outerIndex + 1; innerIndex < concurrentCount; innerIndex++ {
				if serviceHandlers[outerIndex].Handler ==
					serviceHandlers[innerIndex].Handler {
					t.Errorf(
						"HandlersNotIndependent: call %d and %d share handler",
						outerIndex,
						innerIndex,
					)
				}
			}
		}
	})
}
