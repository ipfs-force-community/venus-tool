package service

import (
	"context"

	"github.com/ipfs-force-community/venus-tool/dep"
)

func (s *ServiceImpl) ThreadList(ctx context.Context) ([]*dep.ThreadInfo, error) {
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
