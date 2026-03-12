package Simple_Looger

import (
	"context"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type contextKey string

const (
	ProjectID contextKey = "project-id"
	Uid       contextKey = "uid"
	TraceID   contextKey = "trace-id"
)

type Logger struct {
	*logrus.Entry
}

var (
	projectLogger = Logger{
		logrus.NewEntry(logrus.StandardLogger()),
	}
)

func GetLogger(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}
	fields := logrus.Fields{}

	if projectId := ctx.Value(ProjectID); projectId != nil {
		fields["project_id"] = projectId
	}
	if uid := ctx.Value(Uid); uid != nil {
		fields["uid"] = uid
	}
	if traceId := ctx.Value(TraceID); traceId != nil {
		fields["trace_id"] = traceId
	} else {
		fields["trace_id"] = uuid.New().String()
	}

	return logrus.WithFields(fields)
}

func CtxWithUid(ctx context.Context, uid int) context.Context {
	return context.WithValue(ctx, Uid, uid)
}

func CtxWithProjectId(ctx context.Context, projectId string) context.Context {
	return context.WithValue(ctx, ProjectID, projectId)
}

func CtxWithTraceId(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, TraceID, traceId)
}

func CtxWithLogId(ctx context.Context) context.Context {
	return context.WithValue(ctx, ProjectID, uuid.New().String())
}

func GenCtx(projectId string, uid int) context.Context {
	ctx := CtxWithLogId(context.Background())
	ctx = CtxWithProjectId(ctx, projectId)
	ctx = CtxWithUid(ctx, uid)
	return ctx
}

func GenLogger(projectId string, uid int) *logrus.Entry {
	ctx := GenCtx(projectId, uid)
	return GetLogger(ctx)
}
