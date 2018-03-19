# Using a Local Docker Registry

`localregistry.yaml` is a copy of Kubernete's local registry addon, and is included to make it easier to test
Istio by allowing a developer to push docker images locally rather than to some remote registry.

### Run the registry
To run the local registry in your kubernetes cluster:

```shell
$ kubectl apply -f ./tests/util/localregistry/localregistry.yaml
```

### Expose the Registry

After the registry server is running, expose it locally by executing:

```shell
$ kubectl port-forward --namespace kube-system $POD 5000:5000
```

If you're testing locally with minikube, `$POD` can be set with:

```shell
$ POD=$(kubectl get pods --namespace kube-system -l k8s-app=kube-registry \
  -o template --template '{{range .items}}{{.metadata.name}} {{.status.phase}}{{"\n"}}{{end}}' \
  | grep Running | head -1 | cut -f1 -d' ')
```

### Push Local Images

Build and push the Istio docker images to the local registry, by running:

```shell
$ make docker push HUB=localhost:5000 TAG=latest
```

### Run with Local Images

#### Running E2E Tests
If you're running e2e tests, you can set the test flags:

```shell
$ go test <TEST_PATH> -hub=localhost:5000 -tag=latest
```

#### Hard-coding the Image URL

You can also modify the image URLs in your deployment yaml files directly:

```yaml
image: localhost:5000/<APP_NAME>:latest
```

