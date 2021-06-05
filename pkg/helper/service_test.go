package helper

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	serviceWithoutError := &testService{
		t:      t,
		result: nil,
	}
	serviceWithError := &testService{
		t:      t,
		result: fmt.Errorf("fake error"),
	}

	cases := []struct {
		testDescription string
		startServices   []*testService
		stopServices    []*testService
		expectError     bool
	}{
		{
			testDescription: "No errors on start or stop",
			startServices: []*testService{
				serviceWithoutError,
				serviceWithoutError,
			},
			stopServices: []*testService{
				serviceWithoutError,
				serviceWithoutError,
			},
			expectError: false,
		},
		{
			testDescription: "Errors on all start and stop",
			startServices: []*testService{
				serviceWithError,
				serviceWithError,
			},
			stopServices: []*testService{
				serviceWithError,
				serviceWithError,
			},
			expectError: true,
		},
		{
			testDescription: "One error on stop",
			startServices: []*testService{
				serviceWithoutError,
				serviceWithoutError,
			},
			stopServices: []*testService{
				serviceWithoutError,
				serviceWithError,
			},
			expectError: true,
		},
		{
			testDescription: "One error on start",
			startServices: []*testService{
				serviceWithoutError,
				serviceWithError,
			},
			stopServices: []*testService{
				serviceWithoutError,
				serviceWithoutError,
			},
			expectError: true,
		},
	}

	for i, c := range cases {
		t.Logf("Test iteration %d: %s", i, c.testDescription)

		errGroup, ctx, cancel := NewErrGroupAndContext()
		defer cancel()

		for _, svc := range c.startServices {
			StartService(ctx, errGroup, svc)
		}

		timeoutCtx, timeoutCancel := NewShutdownTimeoutContext()
		defer timeoutCancel()

		for _, svc := range c.stopServices {
			StopService(timeoutCtx, errGroup, svc)
		}

		err := WaitForErrGroup(errGroup)
		if !c.expectError {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}

}

type testService struct {
	t      *testing.T
	result error
}

func (svc *testService) Start(ctx context.Context) error {
	svc.t.Helper()
	return svc.result
}

func (svc *testService) Stop(ctx context.Context) error {
	svc.t.Helper()
	return svc.result
}
