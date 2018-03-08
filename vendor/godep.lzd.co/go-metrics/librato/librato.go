package librato

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
	"time"

	"godep.lzd.co/go-metrics"
)

// a regexp for extracting the unit from time.Duration.String
var unitRegexp = regexp.MustCompile("[^\\d]+$")

// a helper that turns a time.Duration into librato display attributes for timer metrics
func translateTimerAttributes(d time.Duration) (attrs map[string]interface{}) {
	attrs = make(map[string]interface{})
	attrs[DisplayTransform] = fmt.Sprintf("x/%d", int64(d))
	attrs[DisplayUnitsShort] = string(unitRegexp.Find([]byte(d.String())))
	attrs[Aggregate] = true

	return
}

type Reporter struct {
	Email, Token    string
	Source          string
	Interval        time.Duration
	Registry        metrics.Registry
	Percentiles     []float64              // percentiles to report on histogram metrics
	TimerAttributes map[string]interface{} // units in which timers will be displayed
	MetricPrefix    string
	QuietMode       bool
	UseGaugesOnly   bool
	Client          *LibratoClient
}

func NewReporter(r metrics.Registry, d time.Duration, e string, t string, s string, p []float64, u time.Duration) *Reporter {
	return &Reporter{
		Email:           e,
		Token:           t,
		Source:          s,
		Interval:        d,
		Registry:        r,
		Percentiles:     p,
		TimerAttributes: translateTimerAttributes(u),
		Client:          NewLibratoClient(e, t, http.DefaultClient),
	}
}

//NewClientedReporter return Reporter instance
func NewClientedReporter(
	r metrics.Registry,
	d time.Duration,
	e string,
	t string,
	s string,
	p []float64,
	u time.Duration,
	c *LibratoClient,
) *Reporter {
	return &Reporter{
		Email:           e,
		Token:           t,
		Source:          s,
		Interval:        d,
		Registry:        r,
		Percentiles:     p,
		TimerAttributes: translateTimerAttributes(u),
		Client:          c,
	}
}

func Librato(r metrics.Registry, d time.Duration, e string, t string, s string, p []float64, u time.Duration) {
	NewReporter(r, d, e, t, s, p, u).Run()
}

func (self *Reporter) Run() {
	ticker := time.Tick(self.Interval)
	for now := range ticker {

		var metrics Batch
		var err error
		if metrics, err = self.BuildRequest(now, self.Registry); err != nil {
			if self.QuietMode == false {
				log.Printf("ERROR constructing librato request body %s", err)
			}
		}

		if err := self.Client.PostMetrics(metrics); err != nil {
			if self.QuietMode == false {
				log.Printf("ERROR sending metrics to librato %s", err)
			}
		}
	}
}

// calculate sum of squares from data provided by metrics.Histogram
// see http://en.wikipedia.org/wiki/Standard_deviation#Rapid_calculation_methods
func sumSquares(s metrics.Sample) float64 {
	count := float64(s.Count())
	sumSquared := math.Pow(count*s.Mean(), 2)
	sumSquares := math.Pow(count*s.StdDev(), 2) + sumSquared/count
	if math.IsNaN(sumSquares) {
		return 0.0
	}
	return sumSquares
}
func sumSquaresTimer(t metrics.Timer) float64 {
	count := float64(t.Count())
	sumSquared := math.Pow(count*t.Mean(), 2)
	sumSquares := math.Pow(count*t.StdDev(), 2) + sumSquared/count
	if math.IsNaN(sumSquares) {
		return 0.0
	}
	return sumSquares
}

func (self *Reporter) BuildRequest(now time.Time, r metrics.Registry) (snapshot Batch, err error) {
	snapshot = Batch{
		MeasureTime: now.Unix(),
		Source:      self.Source,
	}
	snapshot.MeasureTime = now.Unix()
	snapshot.Gauges = make([]Measurement, 0)
	snapshot.Counters = make([]Measurement, 0)
	histogramGaugeCount := 1 + len(self.Percentiles)
	r.Each(func(name string, metric interface{}) {

		if self.MetricPrefix != "" {
			name = self.MetricPrefix + name
		}
		measurement := Measurement{}
		measurement[Period] = self.Interval.Seconds()

		switch m := metric.(type) {
		case metrics.Counter:
			if m.Count() > 0 {
				if self.UseGaugesOnly {
					measurement[Name] = fmt.Sprintf("%s.%s", name, "sum")
					measurement[Value] = float64(m.Count())
					measurement[Period] = int64(self.Interval.Seconds())
					measurement[Attributes] = map[string]interface{}{
						DisplayUnitsLong:  Operations,
						DisplayUnitsShort: OperationsShort,
						DisplayMin:        "0",
						Aggregate:         true,
						SummarizeFunction: "sum",
					}
					snapshot.Gauges = append(snapshot.Gauges, measurement)
				} else {
					measurement[Name] = fmt.Sprintf("%s.%s", name, "count")
					measurement[Value] = float64(m.Count())
					measurement[Attributes] = map[string]interface{}{
						DisplayUnitsLong:  Operations,
						DisplayUnitsShort: OperationsShort,
						DisplayMin:        "0",
					}
					snapshot.Counters = append(snapshot.Counters, measurement)
				}
			}
		case metrics.Gauge:
			measurement[Name] = name
			measurement[Value] = float64(m.Value())
			snapshot.Gauges = append(snapshot.Gauges, measurement)
		case metrics.GaugeFloat64:
			measurement[Name] = name
			measurement[Value] = float64(m.Value())
			snapshot.Gauges = append(snapshot.Gauges, measurement)
		case metrics.Histogram:
			if m.Count() > 0 {
				gauges := make([]Measurement, histogramGaugeCount, histogramGaugeCount)
				s := m.Sample()
				measurement[Name] = fmt.Sprintf("%s.%s", name, "hist")
				measurement[Count] = uint64(s.Count())
				measurement[Sum] = s.Sum()
				measurement[Max] = float64(s.Max())
				measurement[Min] = float64(s.Min())
				measurement[SumSquares] = sumSquares(s)
				gauges[0] = measurement
				for i, p := range self.Percentiles {
					gauges[i+1] = Measurement{
						Name:   fmt.Sprintf("%s.%.2f", measurement[Name], p),
						Value:  s.Percentile(p),
						Period: measurement[Period],
					}
				}
				snapshot.Gauges = append(snapshot.Gauges, gauges...)
			}
		case metrics.Meter:
			if self.UseGaugesOnly {
				measurement[Name] = fmt.Sprintf("%s.%s", name, "sum")
				measurement[Value] = float64(m.Count())
				measurement[Period] = int64(self.Interval.Seconds())
				measurement[Attributes] = map[string]interface{}{
					DisplayUnitsLong:  Operations,
					DisplayUnitsShort: OperationsShort,
					DisplayMin:        "0",
					Aggregate:         true,
					SummarizeFunction: "sum",
				}
				snapshot.Gauges = append(snapshot.Gauges, measurement)
			} else {
				measurement[Name] = name
				measurement[Value] = float64(m.Count())
				snapshot.Counters = append(snapshot.Counters, measurement)
			}

			snapshot.Gauges = append(snapshot.Gauges,
				Measurement{
					Name:   fmt.Sprintf("%s.%s", name, "1min"),
					Value:  m.Rate1(),
					Period: int64(self.Interval.Seconds()),
					Attributes: map[string]interface{}{
						DisplayUnitsLong:  Operations,
						DisplayUnitsShort: OperationsShort,
						DisplayMin:        "0",
						Aggregate:         true,
						SummarizeFunction: "sum",
					},
				},
			)
		case metrics.Timer:
			if self.UseGaugesOnly {
				measurement[Name] = fmt.Sprintf("%s.%s", name, "sum")
				measurement[Value] = float64(m.Count())
				measurement[Period] = int64(self.Interval.Seconds())
				measurement[Attributes] = map[string]interface{}{
					DisplayUnitsLong:  Operations,
					DisplayUnitsShort: OperationsShort,
					DisplayMin:        "0",
					Aggregate:         true,
					SummarizeFunction: "sum",
				}
				snapshot.Gauges = append(snapshot.Gauges, measurement)
			} else {
				measurement[Name] = name
				measurement[Value] = float64(m.Count())
				snapshot.Counters = append(snapshot.Counters, measurement)
			}
			if m.Count() > 0 {
				libratoName := fmt.Sprintf("%s.%s", name, "timer.mean")
				gauges := make([]Measurement, histogramGaugeCount, histogramGaugeCount)
				gauges[0] = Measurement{
					Name:       libratoName,
					Count:      uint64(m.Count()),
					Sum:        m.Mean() * float64(m.Count()),
					Max:        float64(m.Max()),
					Min:        float64(m.Min()),
					SumSquares: sumSquaresTimer(m),
					Period:     int64(self.Interval.Seconds()),
					Attributes: self.TimerAttributes,
				}
				for i, p := range self.Percentiles {
					gauges[i+1] = Measurement{
						Name:       fmt.Sprintf("%s.timer.%2.0f", name, p*100),
						Value:      m.Percentile(p),
						Period:     int64(self.Interval.Seconds()),
						Attributes: self.TimerAttributes,
					}
				}
				snapshot.Gauges = append(snapshot.Gauges, gauges...)
				snapshot.Gauges = append(snapshot.Gauges,
					Measurement{
						Name:   fmt.Sprintf("%s.%s", name, "rate.1min"),
						Value:  m.Rate1(),
						Period: int64(self.Interval.Seconds()),
						Attributes: map[string]interface{}{
							DisplayUnitsLong:  Operations,
							DisplayUnitsShort: OperationsShort,
							DisplayMin:        "0",
							Aggregate:         true,
							SummarizeFunction: "sum",
						},
					},
				)
			}
		}
	})
	return
}
