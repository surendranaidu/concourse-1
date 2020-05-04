package gardenbackend_test

import (
	"context"
	"errors"
	"testing"

	"github.com/concourse/concourse/worker/containerd/gardenbackend"
	"github.com/concourse/concourse/worker/containerd/gardenbackend/gardenbackendfakes"
	"github.com/concourse/concourse/worker/containerd/containerdfakes"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/runtime/v2/runc/options"
	"github.com/containerd/typeurl"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type KillerSuite struct {
	suite.Suite
	*require.Assertions

	task          *containerdfakes.FakeTask
	processKiller *gardenbackendfakes.FakeProcessKiller
	killer        gardenbackend.Killer
}

func (s *KillerSuite) SetupTest() {
	s.task = new(containerdfakes.FakeTask)
	s.processKiller = new(gardenbackendfakes.FakeProcessKiller)
	s.killer = gardenbackend.NewKiller(
		gardenbackend.WithProcessKiller(s.processKiller),
	)
}

func (s *KillerSuite) TestKillTaskWithNoProcs() {
	s.T().Run("graceful", func(_ *testing.T) {
		err := s.killer.Kill(context.Background(), s.task, gardenbackend.KillGracefully)
		s.NoError(err)

	})

	s.T().Run("ungraceful", func(_ *testing.T) {
		err := s.killer.Kill(context.Background(), s.task, gardenbackend.KillUngracefully)
		s.NoError(err)
	})

	s.Equal(2, s.task.PidsCallCount())
	s.Equal(0, s.task.LoadProcessCallCount())
}

func (s *KillerSuite) TestKillTaskPidsErr() {
	expectedErr := errors.New("pids-err")
	s.task.PidsReturns(nil, expectedErr)

	s.T().Run("graceful", func(_ *testing.T) {
		err := s.killer.Kill(context.Background(), s.task, gardenbackend.KillGracefully)
		s.True(errors.Is(err, expectedErr))
	})

	s.T().Run("ungraceful", func(_ *testing.T) {
		err := s.killer.Kill(context.Background(), s.task, gardenbackend.KillUngracefully)
		s.True(errors.Is(err, expectedErr))
	})
}

func (s *KillerSuite) TestKillTaskWithOnlyInitProc() {
	s.task.PidsReturns([]containerd.ProcessInfo{
		{Pid: 1234, Info: nil}, // the `init` proc returns `info: nil`
	}, nil)

	s.T().Run("graceful", func(_ *testing.T) {
		err := s.killer.Kill(context.Background(), s.task, gardenbackend.KillUngracefully)
		s.NoError(err)
	})

	s.T().Run("ungraceful", func(_ *testing.T) {
		err := s.killer.Kill(context.Background(), s.task, gardenbackend.KillUngracefully)
		s.NoError(err)
	})

	s.Equal(2, s.task.PidsCallCount())
	s.Equal(0, s.task.LoadProcessCallCount())
	s.Equal(0, s.processKiller.KillCallCount())
}

func (s *KillerSuite) TestKillTaskLoadProcessError() {
	procInfo, err := typeurl.MarshalAny(&options.ProcessDetails{
		ExecID: "execution-1",
	})
	s.NoError(err)

	s.task.PidsReturns([]containerd.ProcessInfo{
		{Pid: 123, Info: procInfo},
	}, nil)

	expectedErr := errors.New("load-proc-err")
	s.task.LoadProcessReturns(nil, expectedErr)

	s.T().Run("graceful", func(_ *testing.T) {
		err = s.killer.Kill(context.Background(), s.task, gardenbackend.KillUngracefully)
		s.True(errors.Is(err, expectedErr))
	})

	s.T().Run("ungraceful", func(_ *testing.T) {
		err = s.killer.Kill(context.Background(), s.task, gardenbackend.KillUngracefully)
		s.True(errors.Is(err, expectedErr))
	})
}

func (s *KillerSuite) TestUngracefulKillTaskProcKillError() {
	procInfo, err := typeurl.MarshalAny(&options.ProcessDetails{
		ExecID: "execution-1",
	})
	s.NoError(err)

	s.task.PidsReturns([]containerd.ProcessInfo{
		{Pid: 123, Info: procInfo},
	}, nil)

	expectedErr := errors.New("load-proc-err")
	s.processKiller.KillReturns(expectedErr)

	err = s.killer.Kill(context.Background(), s.task, gardenbackend.KillUngracefully)
	s.True(errors.Is(err, expectedErr))
}

func (s *KillerSuite) TestGracefulKillTaskProcKillGracePeriodTimeoutError() {
	procInfo, err := typeurl.MarshalAny(&options.ProcessDetails{
		ExecID: "execution-1",
	})
	s.NoError(err)

	s.task.PidsReturns([]containerd.ProcessInfo{
		{Pid: 123, Info: procInfo},
	}, nil)

	expectedErr := gardenbackend.ErrGracePeriodTimeout
	s.processKiller.KillReturnsOnCall(0, expectedErr)

	err = s.killer.Kill(context.Background(), s.task, gardenbackend.KillGracefully)
	s.NoError(err)

	s.Equal(2, s.processKiller.KillCallCount())
}

func (s *KillerSuite) TestGracefulKillTaskProcKillUncaughtError() {
	procInfo, err := typeurl.MarshalAny(&options.ProcessDetails{
		ExecID: "execution-1",
	})
	s.NoError(err)

	s.task.PidsReturns([]containerd.ProcessInfo{
		{Pid: 123, Info: procInfo},
	}, nil)

	expectedErr := errors.New("kill-err")
	s.processKiller.KillReturnsOnCall(0, expectedErr)

	err = s.killer.Kill(context.Background(), s.task, gardenbackend.KillGracefully)
	s.True(errors.Is(err, expectedErr))

	s.Equal(1, s.processKiller.KillCallCount())
}

func (s *KillerSuite) TestGracefulKillTaskProcKillErrorOnUngracefulTry() {
	procInfo, err := typeurl.MarshalAny(&options.ProcessDetails{
		ExecID: "execution-1",
	})
	s.NoError(err)

	s.task.PidsReturns([]containerd.ProcessInfo{
		{Pid: 123, Info: procInfo},
	}, nil)

	s.processKiller.KillReturnsOnCall(0, gardenbackend.ErrGracePeriodTimeout)
	expectedErr := errors.New("ungraceful-kill-err")
	s.processKiller.KillReturnsOnCall(1, expectedErr)

	err = s.killer.Kill(context.Background(), s.task, gardenbackend.KillGracefully)
	s.True(errors.Is(err, expectedErr))

	s.Equal(2, s.processKiller.KillCallCount())
}
