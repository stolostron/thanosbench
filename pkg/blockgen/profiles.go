package blockgen

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/timestamp"
	"github.com/prometheus/prometheus/tsdb"
	"github.com/thanos-io/thanos/pkg/block/metadata"
	"github.com/thanos-io/thanos/pkg/model"
	"github.com/thanos-io/thanosbench/pkg/seriesgen"
)

type PlanFn func(ctx context.Context, maxTime model.TimeOrDurationValue, extLset labels.Labels, blockEncoder func(BlockSpec) error) error
type ProfileMap map[string]PlanFn

func (p ProfileMap) Keys() (keys []string) {
	for k := range p {
		keys = append(keys, k)
	}
	return keys
}

var (
	Profiles = ProfileMap{
		// Let's say we have 100 applications, 50 metrics each. All rollout every 1h.
		// This makes 2h block to have 15k series, 8h block 45k, 2d block to have 245k series.
		"realistic-k8s-2d-small": realisticK8s([]time.Duration{
			// Two days, from newest to oldest, in the same way Thanos compactor would do.
			2 * time.Hour,
			2 * time.Hour,
			2 * time.Hour,
			8 * time.Hour,
			8 * time.Hour,
			8 * time.Hour,
			8 * time.Hour,
			8 * time.Hour,
			2 * time.Hour,
		}, 1*time.Hour, 100, 50),
		"realistic-k8s-1w-small": realisticK8s([]time.Duration{
			// One week, from newest to oldest, in the same way Thanos compactor would do.
			2 * time.Hour,
			2 * time.Hour,
			2 * time.Hour,
			8 * time.Hour,
			8 * time.Hour,
			48 * time.Hour,
			48 * time.Hour,
			48 * time.Hour,
			2 * time.Hour,
		}, 1*time.Hour, 100, 50),
		"realistic-k8s-30d-tiny": realisticK8s([]time.Duration{
			// 30 days, from newest to oldest.
			2 * time.Hour,
			2 * time.Hour,
			2 * time.Hour,
			8 * time.Hour,
			176 * time.Hour,
			176 * time.Hour,
			176 * time.Hour,
			176 * time.Hour,
			2 * time.Hour,
		}, 1*time.Hour, 1, 5),
		"realistic-k8s-365d-tiny": realisticK8s([]time.Duration{
			// 1y days, from newest to oldest.
			2 * time.Hour,
			2 * time.Hour,
			2 * time.Hour,
			8 * time.Hour,
			176 * time.Hour,
			176 * time.Hour,
			176 * time.Hour,
			176 * time.Hour,
			67 * 24 * time.Hour,
			67 * 24 * time.Hour,
			67 * 24 * time.Hour,
			67 * 24 * time.Hour,
			67 * 24 * time.Hour,
		}, 1*time.Hour, 1, 5),
		"continuous-1w-small": continuous([]time.Duration{
			// One week, from newest to oldest, in the same way Thanos compactor would do.
			2 * time.Hour,
			2 * time.Hour,
			2 * time.Hour,
			8 * time.Hour,
			8 * time.Hour,
			48 * time.Hour,
			48 * time.Hour,
			48 * time.Hour,
			2 * time.Hour,
			// 10,000 series per block.
		}, 100, 100),
		"continuous-30d-tiny": continuous([]time.Duration{
			// 30 days, from newest to oldest.
			2 * time.Hour,
			2 * time.Hour,
			2 * time.Hour,
			8 * time.Hour,
			176 * time.Hour,
			176 * time.Hour,
			176 * time.Hour,
			176 * time.Hour,
			2 * time.Hour,
		}, 1, 5),
		"continuous-365d-tiny": continuous([]time.Duration{
			// 1y days, from newest to oldest.
			2 * time.Hour,
			2 * time.Hour,
			2 * time.Hour,
			8 * time.Hour,
			176 * time.Hour,
			176 * time.Hour,
			176 * time.Hour,
			176 * time.Hour,
			67 * 24 * time.Hour,
			67 * 24 * time.Hour,
			67 * 24 * time.Hour,
			67 * 24 * time.Hour,
			67 * 24 * time.Hour,
		}, 1, 5),
		"continuous-1w-1series-10000apps": continuous([]time.Duration{
			// One week, from newest to oldest, in the same way Thanos compactor would do.
			2 * time.Hour,
			2 * time.Hour,
			2 * time.Hour,
			8 * time.Hour,
			8 * time.Hour,
			48 * time.Hour,
			48 * time.Hour,
			48 * time.Hour,
			2 * time.Hour,
			// 10,000 series per block.
		}, 10000, 1),
		"cpubench": custom_continuous([]time.Duration{
			// One week, from newest to oldest, in the same way Thanos compactor would do.
			2 * time.Hour,
			2 * time.Hour,
			2 * time.Hour,
			8 * time.Hour,
			48 * time.Hour,
			48 * time.Hour,
			48 * time.Hour,
			192 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			350 * time.Hour,
			// 10,000 series per block.
		}, 4000, []string{"metric1", "metric2", "metric3", "metric4", "metric5"}),
		"sixmonth": custom_continuous([]time.Duration{
			// six months of data, 7800 apps
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
		}, 7800, []string{
			"kube_pod_container_resource_requests:sum",
			"kube_node_status_allocatable",
			"node_cpu_seconds_total",
			"node_namespace_pod_container:container_cpu_usage_seconds_total:sum",
			"kube_pod_owner",
			":node_memory_MemAvailable_bytes:sum",
			"container_memory_rss:sum",
			"instance:node_network_receive_bytes_excluding_lo:rate1m",
			"instance:node_network_transmit_bytes_excluding_lo:rate1m",
			"instance:node_network_receive_drop_excluding_lo:rate1m",
			"instance:node_network_transmit_drop_excluding_lo:rate1m",
		}),
		"sixmonth_memory": custom_continuous([]time.Duration{
			// six months of data, 7800 apps
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
			146 * time.Hour,
		}, 7800, []string{
			"kube_pod_container_resource_requests:sum",
			"kube_node_status_allocatable",
		}),
	}
)

// func profileFactory(profileType string, blockDurationList []int, rolloutInterval int, apps int, metricsPerApp int, customMetrics []string) PlanFn {
// 	var ranges []time.Duration
// 	for _, item := range blockDurationList {
// 		ranges = append(ranges, time.Duration(item)*time.Hour)
// 	}
// 	switch profileType {
// 	case "realisticK8s":
// 		return realisticK8s(ranges, time.Duration(rolloutInterval), apps, metricsPerApp)
// 	case "continuous":
// 		return continuous(ranges, apps, metricsPerApp)
// 	case "custom":
// 		return custom_continuous(ranges, apps, customMetrics)
// 	}
// }

func realisticK8s(ranges []time.Duration, rolloutInterval time.Duration, apps int, metricsPerApp int) PlanFn {
	return func(ctx context.Context, maxTime model.TimeOrDurationValue, extLset labels.Labels, blockEncoder func(BlockSpec) error) error {

		// Align timestamps as Prometheus would do.
		maxt := rangeForTimestamp(maxTime.PrometheusTimestamp(), durToMilis(2*time.Hour))

		// Track "rollouts". In heavy used K8s we have rollouts e.g every hour if not more. Account for that.
		lastRollout := maxt - (durToMilis(rolloutInterval) / 2)

		// All our series are gauges.
		common := SeriesSpec{
			Targets: apps,
			Type:    Gauge,
			Characteristics: seriesgen.Characteristics{
				Max:            200000000,
				Min:            10000000,
				Jitter:         30000000,
				ScrapeInterval: 15 * time.Second,
				ChangeInterval: 1 * time.Hour,
			},
		}

		for _, r := range ranges {
			mint := maxt - durToMilis(r) + 1

			b := BlockSpec{
				Meta: metadata.Meta{
					BlockMeta: tsdb.BlockMeta{
						MaxTime:    maxt,
						MinTime:    mint,
						Compaction: tsdb.BlockMetaCompaction{Level: 1},
						Version:    1,
					},
					Thanos: metadata.Thanos{
						Labels:     extLset.Map(),
						Downsample: metadata.ThanosDownsample{Resolution: 0},
						Source:     "blockgen",
					},
				},
			}
			for {
				if ctx.Err() != nil {
					return ctx.Err()
				}

				smaxt := lastRollout + durToMilis(rolloutInterval)
				if smaxt > maxt {
					smaxt = maxt
				}

				smint := lastRollout
				if smint < mint {
					smint = mint
				}

				for i := 0; i < metricsPerApp; i++ {
					s := common

					s.Labels = labels.Labels{
						// TODO(bwplotka): Use different label for metricPerApp cardinality and stable number.
						{Name: "__name__", Value: fmt.Sprintf("k8s_app_metric%d", i)},
						{Name: "next_rollout_time", Value: timestamp.Time(lastRollout).String()},
					}
					s.MinTime = smint
					s.MaxTime = smaxt
					b.Series = append(b.Series, s)
				}

				if lastRollout <= mint {
					break
				}

				lastRollout -= durToMilis(rolloutInterval)
			}

			if err := blockEncoder(b); err != nil {
				return err
			}
			maxt = mint
		}
		return nil
	}
}

func custom_continuous(ranges []time.Duration, apps int, metrics []string) PlanFn {
	return func(ctx context.Context, maxTime model.TimeOrDurationValue, extLset labels.Labels, blockEncoder func(BlockSpec) error) error {

		// Align timestamps as Prometheus would do.
		maxt := rangeForTimestamp(maxTime.PrometheusTimestamp(), durToMilis(2*time.Hour))

		// All our series are gauges.
		common := SeriesSpec{
			Targets: apps,
			Type:    Gauge,
			Characteristics: seriesgen.Characteristics{
				Max:            200000000,
				Min:            10000000,
				Jitter:         30000000,
				ScrapeInterval: 15 * time.Second,
				ChangeInterval: 1 * time.Hour,
			},
		}

		for _, r := range ranges {
			mint := maxt - durToMilis(r) + 1

			if ctx.Err() != nil {
				return ctx.Err()
			}

			b := BlockSpec{
				Meta: metadata.Meta{
					BlockMeta: tsdb.BlockMeta{
						MaxTime:    maxt,
						MinTime:    mint,
						Compaction: tsdb.BlockMetaCompaction{Level: 1},
						Version:    1,
					},
					Thanos: metadata.Thanos{
						Labels:     extLset.Map(),
						Downsample: metadata.ThanosDownsample{Resolution: 0},
						Source:     "blockgen",
					},
				},
			}

			// append specific metric names
			for _, metric := range metrics {
				s := common

				s.Labels = labels.Labels{
					{Name: "__name__", Value: metric},
				}
				s.MinTime = mint
				s.MaxTime = maxt
				b.Series = append(b.Series, s)
			}
			maxt = mint
		}
		return nil
	}
}

func continuous(ranges []time.Duration, apps int, metricsPerApp int) PlanFn {
	return func(ctx context.Context, maxTime model.TimeOrDurationValue, extLset labels.Labels, blockEncoder func(BlockSpec) error) error {

		// Align timestamps as Prometheus would do.
		maxt := rangeForTimestamp(maxTime.PrometheusTimestamp(), durToMilis(2*time.Hour))

		// All our series are gauges.
		common := SeriesSpec{
			Targets: apps,
			Type:    Gauge,
			Characteristics: seriesgen.Characteristics{
				Max:            200000000,
				Min:            10000000,
				Jitter:         30000000,
				ScrapeInterval: 15 * time.Second,
				ChangeInterval: 1 * time.Hour,
			},
		}

		for _, r := range ranges {
			mint := maxt - durToMilis(r) + 1

			if ctx.Err() != nil {
				return ctx.Err()
			}

			b := BlockSpec{
				Meta: metadata.Meta{
					BlockMeta: tsdb.BlockMeta{
						MaxTime:    maxt,
						MinTime:    mint,
						Compaction: tsdb.BlockMetaCompaction{Level: 1},
						Version:    1,
					},
					Thanos: metadata.Thanos{
						Labels:     extLset.Map(),
						Downsample: metadata.ThanosDownsample{Resolution: 0},
						Source:     "blockgen",
					},
				},
			}
			for i := 0; i < metricsPerApp; i++ {
				s := common

				s.Labels = labels.Labels{
					{Name: "__name__", Value: fmt.Sprintf("continuous_app_metric%d", i)},
				}
				s.MinTime = mint
				s.MaxTime = maxt
				b.Series = append(b.Series, s)
			}

			if err := blockEncoder(b); err != nil {
				return err
			}
			maxt = mint
		}
		return nil
	}
}

func rangeForTimestamp(t int64, width int64) (maxt int64) {
	return (t/width)*width + width
}
