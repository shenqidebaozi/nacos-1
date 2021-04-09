package registry

import (
	"context"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var (
	_ registry.Watcher = (*watcher)(nil)
)

type watcher struct {
	serviceName string
	ctx         context.Context
	cancel      context.CancelFunc
	watchChan   chan bool
	cli         naming_client.INamingClient
}

func newWatcher(ctx context.Context, cli naming_client.INamingClient, serviceName string, groupName string, clusters []string) (*watcher, error) {
	w := &watcher{
		serviceName: serviceName,
		cli:         cli,
		watchChan:   make(chan bool, 1),
	}
	w.ctx, w.cancel = context.WithCancel(ctx)

	e := w.cli.Subscribe(&vo.SubscribeParam{
		ServiceName: serviceName,
		Clusters:    clusters,
		GroupName:   groupName,
		SubscribeCallback: func(services []model.SubscribeService, err error) {
			w.watchChan <- true
		},
	})
	return w, e
}

func (w *watcher) Next() ([]*registry.ServiceInstance, error) {
	for {
		select {
		case <-w.ctx.Done():
			return nil, w.ctx.Err()
		case <-w.watchChan:
		}
		res, err := w.cli.GetService(vo.GetServiceParam{
			ServiceName: w.serviceName,
		})
		if err != nil {
			return nil, err
		}
		var items []*registry.ServiceInstance
		for _, in := range res.Hosts {
			si, e := unmarshal(in)
			if e != nil {
				return nil, err
			}
			items = append(items, si)
		}
		return items, nil
	}
}

func (w *watcher) Stop() error {
	w.cancel()
	//close
	return nil
}
