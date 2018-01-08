istio admin:
    https://istio.io/docs/tasks/telemetry/using-istio-dashboard.html

Steps:
    1.Install istio in the kubernetes cluster and deploy an application;
    2.Install the Prometheus add-on:
        kubectl apply -f prometheus.yaml
    3.Execute the following command:
        kubectl apply -f grafana.yaml
    4.Verify that the service is running:
        kubectl -n istio-system get svc grafana
    5.Open the Istio Dashboard via the Grafana UI:
        kubectl -n istio-system port-forward $(kubectl -n istio-system get pod -l app=grafana -o jsonpath='{.items[0].metadata.name}') 3000:31009 &
    6.Visit http://localhost:31009/dashboard/db/istio-dashboard in web browser


