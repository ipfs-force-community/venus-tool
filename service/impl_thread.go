package service

import (
	"context"
	"fmt"

	"github.com/ipfs-force-community/venus-tool/dep"
)

func (s *ServiceImpl) ThreadList(ctx context.Context) ([]*dep.ThreadInfo, error) {
	if s.Damocles == nil {
		return nil, ErrEmptyDamocles
	}
	var ret []*dep.ThreadInfo
	pingInfos, err := s.Damocles.WorkerPingInfoList(ctx)
	if err != nil {
		return nil, err
	}
	for i := range pingInfos {
		pingInfo := pingInfos[i]
		workerCli, closer, err := dep.NewWorkerClient(ctx, &pingInfo)
		if err != nil {
			log.Warnw("create worker client failed", "error", err, "worker", pingInfo.Info.Name, "addr", pingInfo.Info.Dest)
		}
		defer closer()

		details, err := workerCli.WorkerList()
		if err != nil {
			log.Warnw("get thread detail failed", "error", err, "worker", pingInfo.Info.Name, "addr", pingInfo.Info.Dest)
		}

		for j := range details {
			detail := details[j]
			ret = append(ret, &dep.ThreadInfo{
				WorkerInfo:       &pingInfo.Info,
				WorkerThreadInfo: &detail,
				LastPing:         pingInfo.LastPing,
			})

		}

	}
	return ret, nil
}

func (s *ServiceImpl) ThreadStop(ctx context.Context, req *ThreadStopReq) error {
	workerCli, closer, err := s.getWorkerClient(ctx, req.WorkerName)
	if err != nil {
		return err
	}
	defer closer()
	ok, err := workerCli.WorkerPause(req.Index)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("worker %s pause failed", req.WorkerName)
	}
	return nil
}

func (s *ServiceImpl) ThreadStart(ctx context.Context, req *ThreadStartReq) error {
	workerName, index, state := req.WorkerName, req.Index, req.State
	workerCli, closer, err := s.getWorkerClient(ctx, workerName)
	if err != nil {
		return err
	}
	defer closer()
	ok, err := workerCli.WorkerResume(index, state)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("worker %s start failed", workerName)
	}
	return nil
}

func (s *ServiceImpl) getWorkerClient(ctx context.Context, workerName string) (*dep.WorkerClient, func(), error) {
	if s.Damocles == nil {
		return nil, nil, ErrEmptyDamocles
	}
	pingInfos, err := s.Damocles.WorkerPingInfoList(ctx)
	if err != nil {
		return nil, nil, err
	}

	var target *dep.WorkerPingInfo
	for i := range pingInfos {
		pingInfo := pingInfos[i]
		if pingInfo.Info.Name != workerName {
			continue
		}
		target = &pingInfo
	}
	if target == nil {
		return nil, nil, fmt.Errorf("worker %s not found", workerName)
	}
	return dep.NewWorkerClient(ctx, target)
}
