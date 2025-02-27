package prometheus

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/pipe-cd/pipecd/pkg/app/piped/analysisprovider/metrics"
)

func TestType(t *testing.T) {
	p := Provider{}
	assert.Equal(t, ProviderType, p.Type())
}

func TestProviderQueryPoints(t *testing.T) {
	testcases := []struct {
		name       string
		client     client
		query      string
		queryRange metrics.QueryRange
		want       []metrics.DataPoint
		wantErr    bool
	}{
		{
			name: "query failed",
			client: &fakeClient{
				err: fmt.Errorf("query error"),
			},
			query: "foo",
			queryRange: metrics.QueryRange{
				From: time.Date(2009, time.January, 1, 0, 0, 0, 0, time.UTC),
				To:   time.Date(2009, time.January, 1, 0, 5, 0, 0, time.UTC),
			},
			wantErr: true,
		},
		{
			name: "scalar data point returned",
			client: &fakeClient{
				value: &model.Scalar{Timestamp: model.Time(1600000000), Value: model.SampleValue(0.1)},
			},
			query: "foo",
			queryRange: metrics.QueryRange{
				From: time.Date(2009, time.January, 1, 0, 0, 0, 0, time.UTC),
				To:   time.Date(2009, time.January, 1, 0, 5, 0, 0, time.UTC),
			},
			want: []metrics.DataPoint{
				{Timestamp: 1600000, Value: 0.1},
			},
		},
		{
			name: "vector data points returned",
			client: &fakeClient{
				value: model.Vector([]*model.Sample{
					{
						Timestamp: model.Time(1600000000),
						Value:     model.SampleValue(0.1),
					},
					{
						Timestamp: model.Time(1600001000),
						Value:     model.SampleValue(0.2),
					},
				}),
			},
			query: "foo",
			queryRange: metrics.QueryRange{
				From: time.Date(2009, time.January, 1, 0, 0, 0, 0, time.UTC),
				To:   time.Date(2009, time.January, 1, 0, 5, 0, 0, time.UTC),
			},
			want: []metrics.DataPoint{
				{Timestamp: 1600000, Value: 0.1},
				{Timestamp: 1600001, Value: 0.2},
			},
		},
		{
			name: "matrix data points returned",
			client: &fakeClient{
				value: model.Matrix([]*model.SampleStream{
					{
						Values: []model.SamplePair{
							{
								Timestamp: model.Time(1600000000),
								Value:     model.SampleValue(0.1),
							},
							{
								Timestamp: model.Time(1600001000),
								Value:     model.SampleValue(0.2),
							},
						},
					},
				}),
			},
			query: "foo",
			queryRange: metrics.QueryRange{
				From: time.Date(2009, time.January, 1, 0, 0, 0, 0, time.UTC),
				To:   time.Date(2009, time.January, 1, 0, 5, 0, 0, time.UTC),
			},
			want: []metrics.DataPoint{
				{Timestamp: 1600000, Value: 0.1},
				{Timestamp: 1600001, Value: 0.2},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			provider := &Provider{
				api:     tc.client,
				timeout: defaultTimeout,
				logger:  zap.NewNop(),
			}
			got, err := provider.QueryPoints(context.Background(), tc.query, tc.queryRange)
			assert.Equal(t, tc.wantErr, err != nil)
			assert.Equal(t, tc.want, got)
		})
	}
}
