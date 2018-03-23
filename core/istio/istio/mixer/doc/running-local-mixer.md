# Running Mixer locally

The following command runs Mixer locally using local configuration.
The default configuration contain adapters like `stackdriver` that connect to outside systems. If you do not intend to use the  adapter for
local testing, you should move `mixer/testdata/config/stackdriver.yaml` out of the config directory, otherwise you will see repeated logging of
configuration errors.

The local configuration uses a Kubernetes attribute producing adapter. 
The `KUBECONFIG` environment variable specifies the location of the Kubernetes configuration.

```shell
KUBECONFIG=${HOME}/.kube/config bazel-bin/mixer/cmd/mixs/mixs server --logtostderr --configStoreURL=fs://$(pwd)/mixer/testdata/config -v=4
```

You can also run a simple client to interact with the server:

The following command sends a `check` request to Mixer.
Note that `source.ip` is an ip address specified as 4 `:` separated bytes. 
`192.0.0.2` is encoded as `c0:0:0:2` in the example.

```shell
bazel-bin/mixer/cmd/mixc/mixc check -v 2 --string_attributes destination.service=abc.ns.svc.cluster.local,source.name=myservice,target.port=8080 --stringmap_attributes "request.headers=clnt:abcd;source:abcd,destination.labels=app:ratings,source.labels=version:v2"   --timestamp_attributes request.time="2017-07-04T00:01:10Z" --bytes_attributes source.ip=c0:0:0:2

Check RPC completed successfully. Check status was OK
  Valid use count: 10000, valid duration: 5m0s
```

The following command sends a `report` request to Mixer.
```shell
bazel-bin/mixer/cmd/mixc/mixc report -v 2 --string_attributes destination.service=abc.ns.svc.cluster.local,source.name=myservice,target.port=8080 --stringmap_attributes "request.headers=clnt:abc;source:abcd,destination.labels=app:ratings,source.labels=version:v2"  --int64_attributes response.duration=2003,response.size=1024 --timestamp_attributes  request.time="2017-07-04T00:01:10Z" --bytes_attributes source.ip=c0:0:0:2

Report RPC returned OK
```
