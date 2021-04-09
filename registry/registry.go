package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pkg/errors"
	"net"
	"strconv"
	"strings"
)

var (
	_ registry.Registrar = (*Registry)(nil)
	_ registry.Discovery = (*Registry)(nil)
)

const (
	defPrefix      = "/golang/registry"
	defClusterName = "DEFAULT"
	defGroupName   = "DEFAULT_GROUP"
	defWeight      = 10
)

type options struct {
	prefixPath string
	//ttl        uint64
	weight float64

	clusterName string
	groupName   string
}

type Option func(o *options)

func WithPrefixPath(prefix string) Option {
	return func(o *options) { o.prefixPath = prefix }
}

func WithWeight(weight float64) Option {
	return func(o *options) { o.weight = weight }
}

//func ClusterName(clusterName string) Option {
//	return func(o *options) { o.clusterName = clusterName }
//}
//func GroupName(groupName string) Option {
//	return func(o *options) { o.groupName = groupName }
//}

type Registry struct {
	opt *options
	cli naming_client.INamingClient
}

func New(iClient naming_client.INamingClient, opts ...Option) (r *Registry) {
	opt := &options{
		prefixPath:  defPrefix,
		clusterName: defClusterName,
		groupName:   defGroupName,
		weight:      defWeight,
	}
	for _, option := range opts {
		option(opt)
	}
	return &Registry{
		opt: opt,
		cli: iClient,
	}
}

// like:
// serviceName@http
// serviceName@grpc
func GenServiceName(serviceName, endpointsServerType string) string {
	return serviceName + "@" + endpointsServerType
}
func SplitServiceName(MixServiceName string) (serviceName, endpointsServerType string, err error) {
	s := strings.Split(MixServiceName, "@")
	if len(s) != 2 {
		return "", "", errors.Errorf("[registry]serviceName :%s")
	}
	return s[0], s[1], nil
}

func resolveEndpoints(endpoint string) (srvName, ip string, port uint64, err error) {
	var p string
	var e error

	endpoints := strings.Split(endpoint, "://")
	if len(endpoints) != 2 {
		err = fmt.Errorf("endpoint err Split %v", endpoint)
		return
	}
	srvName = endpoints[0]
	ip, p, err = net.SplitHostPort(endpoints[1])
	if err != nil {
		err = fmt.Errorf("endpoint err SplitHostPort %v,%v", err, endpoint)
		return
	}
	port, e = strconv.ParseUint(p, 10, 64)
	if e != nil {
		err = fmt.Errorf("endpoint point err %v,%v", e, endpoint)
		return
	}
	return
}

func (r *Registry) Register(ctx context.Context, service *registry.ServiceInstance) error {
	endpoints := service.Endpoints
	for _, endpoint := range endpoints {
		st, ip, port, err := resolveEndpoints(endpoint)
		if err != nil {
			return err
		}

		//keep only the current serverType endpoint
		service.Endpoints = []string{endpoint}
		b, err := json.Marshal(service)
		if err != nil {
			return nil
		}
		service.Metadata[MdColumnSrvInstance] = string(b)

		_, e := r.cli.RegisterInstance(vo.RegisterInstanceParam{
			Ip:          ip,
			Port:        port,
			ServiceName: GenServiceName(service.Name, st),
			Weight:      r.opt.weight,
			Enable:      true,
			Healthy:     true,
			Ephemeral:   true,
			Metadata:    service.Metadata,
			ClusterName: r.opt.clusterName,
			GroupName:   r.opt.groupName,
		})
		if e != nil {
			return fmt.Errorf("RegisterInstance err %v,%v", e, endpoint)
		}
	}
	return nil
}

func (r *Registry) Deregister(ctx context.Context, service *registry.ServiceInstance) error {
	var err error
	for _, endpoint := range service.Endpoints {
		st, ip, port, err := resolveEndpoints(endpoint)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("resolveEndpoints(endpoint) endpoint:%v", endpoint))
			continue
		}
		_, e := r.cli.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          ip,
			Port:        port,
			ServiceName: GenServiceName(service.Name, st),
		})
		if e != nil {
			err = errors.Wrap(e, fmt.Sprintf("DeregisterInstance ip:%v,port:%v,service-name:%v", ip, port, GenServiceName(service.Name, st)))
		}
	}
	return err
}

func (r *Registry) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	groupName := defGroupName
	clusters := []string{defClusterName}
	return newWatcher(ctx, r.cli, serviceName, groupName, clusters)
}

func (r *Registry) GetService(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	res, err := r.cli.GetService(vo.GetServiceParam{
		ServiceName: serviceName,
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
