# Istio Addons

This directory contains components that are not a part of core Istio,
but are built and included with the Istio release. Files that are to
be compiled and/or be built into Docker images go here. Install files
still go under [`/install`](/install).

### [`grafana`](grafana)

Files for Istio Grafana docker image. See
[docs](https://istio.io/docs/tasks/telemetry/using-istio-dashboard.html)
for usage.


### [`servicegraph`](servicegraph)

Code and docker image files for Servicegraph. See
[docs](https://istio.io/docs/tasks/telemetry/servicegraph.html) for
usage.
