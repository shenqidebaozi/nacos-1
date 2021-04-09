package registry

import (
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/nacos-group/nacos-sdk-go/model"
)

const (
	MdColumnSrvInstance = "serviceInstance"
)

func marshal(si *registry.ServiceInstance) (string, error) {
	data, err := json.Marshal(si)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func unmarshal(i model.Instance) (si *registry.ServiceInstance, err error) {
	if i.Metadata != nil {
		if str, ok := i.Metadata[MdColumnSrvInstance]; ok {
			err = json.Unmarshal([]byte(str), &si)
			return
		}
	}

	//reg by other system
	var endpoints []string
	//def endpointsServerType is http ,
	var endpointsServerType = "http"
	//If you want to change the protocol,
	//in other sys,register the service name like:
	//srvName : user@grpc
	//srvName : user@http
	var serviceName = i.ServiceName
	sn, t, e := SplitServiceName(serviceName)
	if e == nil {
		serviceName = sn
		endpointsServerType = t
	}

	endpoints = []string{fmt.Sprint("%s://%s:%s", endpointsServerType, i.Ip, i.Port)}
	si = &registry.ServiceInstance{
		//ID:        id,
		Name: serviceName,
		//Version:   version,
		Metadata:  i.Metadata,
		Endpoints: endpoints,
	}
	return
}
