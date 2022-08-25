package taskmng_test

import (
	"context"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/wmentor/taskmng"
	"github.com/wmentor/taskmng/testdata"
)

const (
	defaultWorkerPoolSize = 16
	defaultTaskQueueSize  = 1024
)

func TestNew_Success(t *testing.T) {
	t.Parallel()

	mng, err := taskmng.New(
		taskmng.WithContext(context.TODO()),
		taskmng.WithWorkerPoolSize(defaultWorkerPoolSize),
		taskmng.WithTaskQueueSize(defaultTaskQueueSize),
	)
	require.NoError(t, err, "New failed")
	require.NotNil(t, mng, "Taskmanager object is nil")
	mng.Stop(true)

	mng, err = taskmng.New(
		taskmng.WithContext(context.TODO()),
		taskmng.WithWorkerPoolSize(defaultWorkerPoolSize),
		taskmng.WithTaskQueueSize(defaultTaskQueueSize),
	)
	require.NoError(t, err, "New failed")
	require.NotNil(t, mng, "Taskmanager object is nil")
	defer mng.Stop(false)
}

func TestProcessTask(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var notifyCounter int64
	var errCounter int64

	callbackKey := "callback"

	mng, err := taskmng.New(
		taskmng.WithContext(context.TODO()),
		taskmng.WithWorkerPoolSize(defaultWorkerPoolSize),
		taskmng.WithTaskQueueSize(defaultTaskQueueSize),
		taskmng.WithEventCallback(callbackKey, func(_ context.Context, _ string) {
			atomic.AddInt64(&notifyCounter, 1)
		}),
		taskmng.WithErrorCallback(func(_ context.Context, _ string) {
			atomic.AddInt64(&errCounter, 1)
		}),
	)
	require.NoError(t, err, "New failed")
	require.NotNil(t, mng, "Taskmanager object is nil")

	var runCounter int64

	limit := 100

	task := testdata.NewMockTask(ctrl)
	task.EXPECT().Exec(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, publisher taskmng.EventPublisher) error {
		val := atomic.AddInt64(&runCounter, 1)
		if val%5 == 0 {
			publisher.PublishEvent(ctx, callbackKey, strconv.FormatInt(val, 10))
		}
		if val%10 == 0 {
			return errors.New("some error")
		}
		return nil
	}).MaxTimes(limit)

	for i := 0; i < limit; i++ {
		mng.AddTask(task)
	}

	mng.Stop(false)

	require.Equal(t, int64(limit), runCounter, "invalid call time")
	require.Equal(t, int64(limit/10), errCounter, "invalid error counter")
	require.Equal(t, int64(limit/5), notifyCounter, "invalid error counter")
}
