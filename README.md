##kratosv2 nacos

---

**example**
````
package server

import (
    "github.com/nacos-group/nacos-sdk-go/clients"
    "github.com/nacos-group/nacos-sdk-go/clients/naming_client"
    "github.com/nacos-group/nacos-sdk-go/common/constant"
    "github.com/nacos-group/nacos-sdk-go/vo"
    ".../internal/conf"
)

// NewHTTPServer new a HTTP server.
func NewNacosServer(c *conf.Data) (naming_client.INamingClient, error) {
    sc := []constant.ServerConfig{
        *constant.NewServerConfig(c.Nacos.Host, c.Nacos.Port),
    }

    cc := constant.ClientConfig{
        NamespaceId:         c.Nacos.NamespaceID,
        TimeoutMs:           5000,
        NotLoadCacheAtStart: true,
        LogDir:              c.Nacos.LogDir,
        CacheDir:            c.Nacos.CacheDir,
        RotateTime:          "1h",
        MaxAge:              3,
        LogLevel:            "debug",
    }

    // a more graceful way to create naming client
    client, err := clients.NewNamingClient(
        vo.NacosClientParam{
            ClientConfig:  &cc,
            ServerConfigs: sc,
        },
    )

    return client, err
}
````