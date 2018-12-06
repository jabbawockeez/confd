package result

import (
    "context"
    // "fmt"
    //"go.etcd.io/etcd/clientv3"
    "github.com/coreos/etcd/clientv3"
    "time"

    "github.com/kelseyhightower/confd/log"
)


var ec *clientv3.Client

func InitEtcdClient(BackendNodes []string) {
    var err error
    ec, err = clientv3.New(clientv3.Config{
        Endpoints:   BackendNodes,
        DialTimeout: 5 * time.Second,
    })
    if err != nil {
        panic(err)
    }
}

func WriteEtcd(key, msg string) {
    log.Info("%v----", msg)
    ec.Put(context.Background(), key, msg)
    // fmt.Println(r.Header, r.PrevKv, err)
}

