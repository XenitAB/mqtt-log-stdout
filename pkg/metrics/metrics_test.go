package metrics

import (
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
	h "github.com/xenitab/mqtt-log-stdout/pkg/helper"
	"github.com/xenitab/mqtt-log-stdout/pkg/status"
	"go.uber.org/goleak"
)

func TestStart(t *testing.T) {
	errGroup, ctx, cancel := h.NewErrGroupAndContext()
	defer cancel()

	statusClient := newFakeStatusClient()

	opts := Options{
		Address:      "0.0.0.0",
		Port:         8080,
		StatusClient: statusClient,
	}
	metricsServer := NewServer(opts)

	h.StartService(ctx, errGroup, metricsServer)

	fakeCounter := promauto.NewCounter(prometheus.CounterOpts{
		Name: "fake_counter",
		Help: "fake counter",
	})

	numberOfWorkers := 10
	messagesPerWorker := 200
	expectedMessageCount := messagesPerWorker * numberOfWorkers
	incrementerErrGroup, _, _ := h.NewErrGroupAndContext()

	for w := 0; w < numberOfWorkers; w++ {
		incrementerErrGroup.Go(func() error {
			for i := 0; i < messagesPerWorker; i++ {
				fakeCounter.Inc()
			}

			return nil
		})
	}

	err := h.WaitForErrGroup(incrementerErrGroup)
	require.NoError(t, err)

	metrics := testGetPrometheusMetrics(t, "http://localhost:8080/metrics")

	cancel()

	timeoutCtx, timeoutCancel := h.NewShutdownTimeoutContext()
	defer timeoutCancel()

	h.StopService(timeoutCtx, errGroup, metricsServer)

	err = h.WaitForErrGroup(errGroup)
	require.NoError(t, err)

	messageCount := int(*metrics["fake_counter"].Metric[0].Counter.Value)
	require.Equal(t, expectedMessageCount, messageCount)
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type fakeStatus struct{}

func newFakeStatusClient() status.Client {
	return &fakeStatus{}
}

func (s *fakeStatus) Print(m string, e error) {}

func testGetPrometheusMetrics(t *testing.T, url string) map[string]*dto.MetricFamily {
	t.Helper()

	res, err := http.Get(url)
	require.NoError(t, err)

	body := res.Body
	defer body.Close()

	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(body)
	require.NoError(t, err)

	return mf
}
