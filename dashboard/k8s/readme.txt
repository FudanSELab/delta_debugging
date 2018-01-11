
k8s dashboard admin:		
    https://github.com/kubernetes/dashboard

First,you need to install the heapster in the cluster. The guidence of installing the heapster can be found in the following links:
    https://github.com/kubernetes/heapster
    https://github.com/kubernetes/heapster/blob/master/docs/influxdb.md

Then, install the dashboard with metrics information:
    https://github.com/kubernetes/dashboard/wiki/Getting-started
    (Notice: Every time you want to rebuild and up the dashboard, you need to first delete the "node_modules" directory. And then reinstall the dependencies by the command "npm i --unsafe-perm")





