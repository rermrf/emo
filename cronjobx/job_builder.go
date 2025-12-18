package cronjobx

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rermrf/emo/logger"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// CronJobBuilder 使用summaryVec对不同的job进行监控
type CronJobBuilder struct {
	l      logger.Logger
	p      *prometheus.SummaryVec
	tracer trace.Tracer
}

// NewCronJobBuilder 创建一个CronJobBuilder
// TODO 抽离创建prometheus.SummaryVec的逻辑
func NewCronJobBuilder(l logger.Logger) *CronJobBuilder {
	opt := prometheus.SummaryOpts{
		Namespace: "emoji",
		Subsystem: "webook",
		Name:      "cron_job",
		Help:      "统计定时任务的执行情况",
	}
	p := prometheus.NewSummaryVec(opt, []string{"name", "success"})
	prometheus.MustRegister(p)
	return &CronJobBuilder{
		l:      l,
		p:      p,
		tracer: otel.GetTracerProvider().Tracer("webook/bff/job/job_builder.go"),
	}
}

func (b *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobFuncAdapter(func() error {
		ctx, span := b.tracer.Start(context.Background(), name)
		defer span.End()
		start := time.Now()
		b.l.Info("任务开始", logger.String("job", name))
		var success bool
		defer func() {
			b.l.Info("任务结束", logger.String("job", name))
			duration := time.Since(start)
			b.p.WithLabelValues(name, strconv.FormatBool(success)).Observe(float64(duration))
		}()
		err := job.Run(ctx)
		success = err == nil
		if err != nil {
			span.RecordError(err)
			b.l.Error("运行任务失败", logger.String("job", job.Name()), logger.Error(err))
		}
		return nil
	})
}

type cronJobFuncAdapter func() error

func (f cronJobFuncAdapter) Run() {
	_ = f()
}
