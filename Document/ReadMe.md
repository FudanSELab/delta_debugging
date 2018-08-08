# Set Up Delta Debugging

We strongly do not recommend to re-run the delta debugging experiment. 
Because we have many hard code in this project and configuration is complex.
And we may forget some details.
This manual is only for reference.

## Step 1 - Prepare
You need at least 4 VMs.
1 of 4 VMs to setup delta debugging system.
3 of 4 VMs to build at least one K8S cluster.

## Step 2 - Config

For cscontrol/delta-backend
Config the cluster information in delta-backend/src/main/resources/application.yml
You could name you k8s cluster name as "cluster1".

For dashboard/apiserver
Config the cluster information in dashboard/apiserver/src/main/resources/application.yml.
The token of the k8s cluster is collect using the following instructions:
    kubectl create -f admin-token.yml
    kubectl get secret -n kube-systemm|grep admin
    kubectl describe secret admin-token-??? -n kube-system

For testmgr/test-backend
Put the test-case that you want to run in testmgr/test-backend/src/main/java/cluster1.
Modify testmgr/test-backend/src/main/resources/docker-testConfig.json and add the testcase name that you put in last step into "Delta Test".

## Step 3 - Setup Delta-Debugging-Project
You could move the whole project to 1 of 4VMs and use the following instructions to setup Delta_Debugging:
    mvn clean package
    docker-compose build
    docker-compose up
    
## Step 4 - Setup Train-Ticket system.
Deploy train-ticket system on your k8s cluster.
If you want to do request-sequence delta, you will use Istio to deploy you Train-Ticket-System with injected sidecar.

## Step 5 - Start delta debugging
Visit your-server-address:8083 to start delta debugging.