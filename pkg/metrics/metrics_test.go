package metrics

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/xenitab/mqtt-log-stdout/pkg/status"
	"go.uber.org/goleak"
	"golang.org/x/sync/errgroup"
)

func TestStart(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	statusClient := newFakeStatusClient()

	opts := Options{
		Address:      "0.0.0.0",
		Port:         8080,
		StatusClient: statusClient,
	}
	metricsServer := NewServer(opts)

	g.Go(func() error {
		err := metricsServer.Start(ctx)
		if err != nil {
			t.Errorf("Expected err to be nil: %q", err)
			return err
		}

		return nil
	})

	fakeCounter := promauto.NewCounter(prometheus.CounterOpts{
		Name: "fake_counter",
		Help: "fake counter",
	})

	expectedMessageCount := 200
	for i := 0; i < expectedMessageCount; i++ {
		fakeCounter.Inc()
	}

	metrics, err := getPrometheusMetrics("http://localhost:8080/metrics")
	if err != nil {
		t.Errorf("Expected err to be nil: %q", err)
	}

	cancel()

	timeoutCtx, timeoutCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer timeoutCancel()

	g.Go(func() error {
		return metricsServer.Stop(timeoutCtx)
	})

	if err := g.Wait(); err != nil {
		t.Errorf("Expected err to be nil: %q", err)
	}

	messageCount := int(*metrics["fake_counter"].Metric[0].Counter.Value)

	if messageCount != expectedMessageCount {
		t.Errorf("Expected message count was %d but received: %d", expectedMessageCount, messageCount)
	}
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type fakeStatus struct{}

func newFakeStatusClient() status.Client {
	return &fakeStatus{}
}

func (s *fakeStatus) Print(m string, e error) {}

func getPrometheusMetrics(url string) (map[string]*dto.MetricFamily, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body := res.Body
	defer body.Close()

	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(body)
	if err != nil {
		return nil, err
	}
	return mf, nil
}
