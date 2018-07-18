package apiserver.service;

import apiserver.async.AsyncTask;
import apiserver.bean.*;
import apiserver.request.*;
import apiserver.response.*;
import apiserver.util.Cluster;
import apiserver.util.MyConfig;
import apiserver.util.RemoteExecuteCommand;
import com.alibaba.fastjson.JSON;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import java.io.*;
import java.util.*;

@Service
public class ApiServiceImpl implements ApiService {

    private final String NAMESPACE = "default";

    @Autowired
    private AsyncTask asyncTask;

    @Autowired
    private MyConfig myConfig;

    public boolean flag = false;

    //Return all the clusters able to control
    @Override
    public GetClustersResponse getClusters() {
        GetClustersResponse response = new GetClustersResponse();
        response.setStatus(false);
        response.setMessage("There is no any clusters now.");
        response.setClusters(null);
        if(myConfig.getClusters().size() > 0){
            response.setStatus(true);
            response.setMessage("Successfully to get the clusters information!");
            response.setClusters(myConfig.getClusters());
        }
        return response;
    }

    @Override
    public SetUnsetServiceRequestSuspendResponse setServiceRequestSuspend(SetUnsetServiceRequestSuspendRequest setUnsetServiceRequestSuspendRequest){
        String svcName = setUnsetServiceRequestSuspendRequest.getSvc();
        Cluster cluster = getClusterByName(setUnsetServiceRequestSuspendRequest.getClusterName());
        System.out.println(String.format("The cluster to operate is [%s]", cluster.getName()));
        String executeResult = doSetServiceRequestSuspend(svcName,cluster);
        System.out.println(executeResult);
        boolean status = (executeResult != null);
        SetUnsetServiceRequestSuspendResponse response = new SetUnsetServiceRequestSuspendResponse(status,executeResult);
        return response;
    }

    private String doSetServiceRequestSuspend(String svcName,Cluster cluster){
        String svcLongDelayFilePath = "rule-long-" + svcName + ".yml";
        RemoteExecuteCommand rec = new RemoteExecuteCommand(cluster.getMasterIp(), cluster.getUsername(),cluster.getPasswd());
        rec.modifyFile(svcLongDelayFilePath,svcName);
        String serLongDelayRequest = "kubectl apply -f " + svcLongDelayFilePath;
        //Execute the script
        String executeResult = rec.execute("export KUBECONFIG=/etc/kubernetes/admin.conf;" + serLongDelayRequest);
        return executeResult;
    }

    @Override
    public SetUnsetServiceRequestSuspendResponse unsetServiceRequestSuspend(SetUnsetServiceRequestSuspendRequest setUnsetServiceRequestSuspendRequest){
        Cluster cluster = getClusterByName(setUnsetServiceRequestSuspendRequest.getClusterName());
        System.out.println(String.format("The cluster to operate is [%s]", cluster.getName()));
        String svcName = setUnsetServiceRequestSuspendRequest.getSvc();
        String executeResult = doUnsetServiceRequestSuspend(svcName,cluster);
        System.out.println(executeResult);
        boolean status = (executeResult != null);
        SetUnsetServiceRequestSuspendResponse response = new SetUnsetServiceRequestSuspendResponse(status,executeResult);
        return response;
    }

    private String doUnsetServiceRequestSuspend(String svcName,Cluster cluster){
        String svcLongDelayFilePath = "rule-long-" + svcName + ".yml";
        String serLongDelayRequest = "kubectl delete -f " + svcLongDelayFilePath;
        RemoteExecuteCommand rec = new RemoteExecuteCommand(cluster.getMasterIp(), cluster.getUsername(),cluster.getPasswd());
        //执行脚本
        return rec.execute("export KUBECONFIG=/etc/kubernetes/admin.conf;" + serLongDelayRequest);
    }


    @Override
    public SetUnsetServiceRequestSuspendResponse setServiceRequestSuspendWithSource(SetUnsetServiceRequestSuspendRequest setUnsetServiceRequestSuspendRequest){
        Cluster cluster = getClusterByName(setUnsetServiceRequestSuspendRequest.getClusterName());
        System.out.println(String.format("The cluster to operate is [%s]", cluster.getName()));
        String svcName = setUnsetServiceRequestSuspendRequest.getSvc();
        String sourceSvcName = setUnsetServiceRequestSuspendRequest.getSourceSvcName();
        String executeResult = doSetServiceRequestSuspendWithSourceFile(svcName,sourceSvcName,cluster);
        System.out.println(executeResult);
        boolean status = (executeResult != null);
        SetUnsetServiceRequestSuspendResponse response = new SetUnsetServiceRequestSuspendResponse(status,executeResult);
        return response;

    }

    private String doSetServiceRequestSuspendWithSourceFile(String svcName, String sourceSvcName,Cluster cluster){
        String svcLongDelayFilePath = "rule-long-" + svcName + "-to-" + sourceSvcName + ".yml";
        RemoteExecuteCommand rec = new RemoteExecuteCommand(cluster.getMasterIp(), cluster.getUsername(),cluster.getPasswd());
        rec.modifyFileWithSourceSvcName(svcLongDelayFilePath,svcName,sourceSvcName);
        String serLongDelayRequest = "kubectl apply -f " + svcLongDelayFilePath;
        //Execute the script
        String executeResult = rec.execute("export KUBECONFIG=/etc/kubernetes/admin.conf;" + serLongDelayRequest);
        return executeResult;
    }

    @Override
    public SetUnsetServiceRequestSuspendResponse unsetServiceRequestSuspendWithSource(
            SetUnsetServiceRequestSuspendRequest request){
        Cluster cluster = getClusterByName(request.getClusterName());
        System.out.println(String.format("The cluster to operate is [%s]", cluster.getName()));
        String svcName = request.getSvc();
        String executeResult = doUnsetServiceRequestSuspendWithSource(svcName,cluster,request.getSourceSvcName());
        System.out.println(executeResult);
        boolean status = (executeResult != null);
        SetUnsetServiceRequestSuspendResponse response = new SetUnsetServiceRequestSuspendResponse(status,executeResult);
        return response;
    }

    private String doUnsetServiceRequestSuspendWithSource(String svcName,Cluster cluster,String sourceSvcName){
        String svcLongDelayFilePath = "rule-long-" + svcName + "-to-" + sourceSvcName + ".yml";
        String serLongDelayRequest = "kubectl delete -f " + svcLongDelayFilePath;
        RemoteExecuteCommand rec = new RemoteExecuteCommand(cluster.getMasterIp(), cluster.getUsername(),cluster.getPasswd());
        //执行脚本
        return rec.execute("export KUBECONFIG=/etc/kubernetes/admin.conf;" + serLongDelayRequest);
    }

    @Override
    public SetAsyncRequestSequenceResponse unsuspendAllRequest(SetAsyncRequestSequenceRequestWithSource request){

        flag = false;

        Cluster cluster = getClusterByName(request.getClusterName());
        System.out.println(String.format("The cluster to operate is [%s]", cluster.getName()));
        ArrayList<String> svcList = request.getSvcList();
        String resultStr = "";
        for(String svcName: svcList){
            String tempStr = doUnsetServiceRequestSuspend(svcName,cluster);
            resultStr += tempStr;
            System.out.println(tempStr);
        }
        return new SetAsyncRequestSequenceResponse(true,resultStr);
    }

    @Override
    public SetAsyncRequestSequenceResponse setAsyncRequestsSequence(SetAsyncRequestSequenceRequest setAsyncRequestSequenceRequest){
        Cluster cluster = getClusterByName(setAsyncRequestSequenceRequest.getClusterName());
        System.out.println(String.format("The cluster to operate is [%s]", cluster.getName()));
        ArrayList<String> svcList = setAsyncRequestSequenceRequest.getSvcList();
        for(int i = 0;i < svcList.size(); i++){
            String svcName = svcList.get(i);
            System.out.println("[=====]Release " + svcName + ": " + doUnsetServiceRequestSuspend(svcName,cluster));
            //waitForComplete是阻塞式的 会一直等待直到请求返回
            if(waitForComplete(svcName, cluster) == true) {
                System.out.println("[===== Complete =====] " + svcName);
            }
        }
        SetAsyncRequestSequenceResponse response = new SetAsyncRequestSequenceResponse(true," setAsyncRequestsSequence Complete");
        return response;
    }

    @Override
    public SetAsyncRequestSequenceResponse setAsyncRequestsSequenceWithSource(SetAsyncRequestSequenceRequestWithSource request){
        Cluster cluster = getClusterByName(request.getClusterName());
        System.out.println(String.format("The cluster to operate is [%s]", cluster.getName()));
        ArrayList<String> svcList = request.getSvcList();
        String srcName = request.getSourceName();
        String sourceSvcName = request.getSourceName();
        for(int i = 0;i < svcList.size(); i++){
            String svcName = svcList.get(i);
            System.out.println("[=====]Release " + svcName + ": " + doUnsetServiceRequestSuspendWithSource(svcName, cluster,sourceSvcName));
            //waitForComplete是阻塞式的 会一直等待直到请求返回
            if(waitForCompleteWithSource(srcName,svcName,cluster) == true) {
                System.out.println("[===== Complete =====] " + svcName);
            }
        }
        System.out.println("[===== Congratulations! All Complete! =====]");
        SetAsyncRequestSequenceResponse response = new SetAsyncRequestSequenceResponse(true," setAsyncRequestsSequence Complete");
        return response;
    }

    @Override
    public SetAsyncRequestSequenceResponse setAsyncRequestSequenceWithSrcCombineWithFullSuspend(SetAsyncRequestSequenceRequestWithSource request){
        Cluster cluster = getClusterByName(request.getClusterName());
        System.out.println(String.format("The cluster to operate is [%s]", cluster.getName()));
        String str = "";
        for(int i = 0;i < request.getSvcList().size();i++){
            String executeResult = doSetServiceRequestSuspendWithSourceFile(request.getSvcList().get(i),request.getSourceName(), cluster);
            str += executeResult;
            System.out.println(executeResult);
        }
        try{
            asyncTask.doAsync(request);
        }catch (Exception e){
            e.printStackTrace();
        }
        System.out.println("[=====]setAsyncRequestSequenceWithSrcCombineWithFullSuspend返回");
        return new SetAsyncRequestSequenceResponse(true,str);
        //return setAsyncRequestsSequenceWithSource(request);
    }

    @Override
    public SetAsyncRequestSequenceResponse controlSequenceAndMaintainIt(SetAsyncRequestSequenceRequestWithSource request){

        flag = true;

        Cluster cluster = getClusterByName(request.getClusterName());
        System.out.println(String.format("The cluster to operate is [%s]", cluster.getName()));
        //丢出一个异步线程，进行控制和解除，以及重新锁定
        try{
            asyncTask.doAsyncWithMaintainSequence(request);
        }catch (Exception e){
            e.printStackTrace();
        }
        //返回
        System.out.println("[=====]setAsyncRequestSequenceWithSrcCombineWithFullSuspend返回");
        return new SetAsyncRequestSequenceResponse(true,"Just return but we do not gaurantee success.");
    }

    @Override
    public SetAsyncRequestSequenceResponse setAsyncRequestSequenceWithSrcCombineWithFullSuspendWithMaintainSequence(SetAsyncRequestSequenceRequestWithSource request){
        Cluster cluster = getClusterByName(request.getClusterName());
        System.out.println(String.format("The cluster to operate is [%s]", cluster.getName()));
        ArrayList<String> svcList = request.getSvcList();
        String srcName = request.getSourceName();
        String sourceSvcName = request.getSourceName();
        while(flag){
            System.out.println("[===] This is a NEW TURN");
            //开始锁定请求顺序
            String str = "";
            for(int i = 0;i < request.getSvcList().size();i++){
                String executeResult = doSetServiceRequestSuspendWithSourceFile(
                        request.getSvcList().get(i),request.getSourceName(), cluster);
                str += executeResult;
                System.out.println(executeResult);
            }
            //然后开始逐步释放
            for(int i = 0;i < svcList.size(); i++){
                String svcName = svcList.get(i);
                System.out.println("[=====]Release " + svcName + ": " + doUnsetServiceRequestSuspendWithSource(svcName, cluster,sourceSvcName));
                //waitForComplete是阻塞式的 会一直等待直到请求返回
                if(waitForCompleteWithSource(srcName,svcName,cluster) == true) {
                    System.out.println("[===== Complete =====] " + svcName);
                }
            }
            System.out.println("[===== Congratulations! All Complete! =====]");
        }
        SetAsyncRequestSequenceResponse response = new SetAsyncRequestSequenceResponse(true," setAsyncRequestsSequence Complete");
        return response;
    }



    private boolean waitForComplete(String svcName, Cluster cluster){
        //根据svc的名称，获取svc下的所有pod
        GetPodsListResponse podsListResponse = getPodsList("default",cluster);
        ArrayList<PodInfo> podInfoList = new ArrayList<>(podsListResponse.getPods());
        ArrayList<PodInfo> targetPodInfoList = new ArrayList<>();
        for(PodInfo podInfo : podInfoList){
            if(podInfo.getName().contains(svcName)){
                System.out.println("[=====] We are now checking useful POD-NAME:" + podInfo.getName());
                targetPodInfoList.add(podInfo);
            }else{
                //do nothing
            }
        }
        boolean isRequestComplete = false;
        while(isRequestComplete == false){
            //每间隔20秒，获取一次pods的日志。注意是pod下的istio-proxy的日志
            try{
                Thread.sleep(20000);
            }catch (InterruptedException e){
                e.printStackTrace();
            }
            //获取各个pod的日志，并截取最后四条
            for(PodInfo podInfo : targetPodInfoList) {
                if(isRequestComplete == true){
                    break;
                }
                System.out.println("[=====] We are now checking POD-LOG:" + podInfo.getName());
                String podLog = getPodLog(podInfo.getName(),"istio-proxy",cluster);
                String[] logsFormatted = podLog.split("\n");
                ArrayList<String> arrayList = new ArrayList<>(Arrays.asList(logsFormatted));
                ArrayList<String> lastSeveralLogs = new ArrayList<>(arrayList.subList(arrayList.size() - 10,arrayList.size()));
                System.out.println("[=====]读取到的日志数量为：" + arrayList.size());
                System.out.println("[=====]截取的日志数量为：" + arrayList.size());
                for(String logStr : lastSeveralLogs) {
                    System.out.println("[=======]Log Line:" + logStr);
                    //并检查日志     response-code 200 && svcName && 接口名称
                    //检查log是否合规
                    if(checkLogCanComfirmRequestComplete(logStr,svcName)){
                        isRequestComplete = true;
                        break;
                    }else{
                        isRequestComplete = false;
                    }
                }
            }
        }
        return isRequestComplete;
    }

    private boolean waitForCompleteWithSource(String srcName, String svcName, Cluster cluster){
        //根据svc的名称，获取svc下的所有pod
        boolean isRequestComplete = false;
        int count = 0;
        GetPodsListResponse podsListResponse = getPodsList("default", cluster);
        ArrayList<PodInfo> podInfoList = new ArrayList<>(podsListResponse.getPods());
        ArrayList<PodInfo> targetPodInfoList = new ArrayList<>();
        for(PodInfo podInfo : podInfoList){
            if(podInfo.getName().contains(srcName)){ //寻找source pod的日志，在source pod里看看有没有svcName
                System.out.println("[=====] We are now checking useful POD-NAME:" + podInfo.getName());
                targetPodInfoList.add(podInfo);
            }else{
                //do nothing
            }
        }
        while(isRequestComplete == false){
            //每间隔 0秒，获取一次pods的日志。注意是pod下的istio-proxy的日志
            try{
                Thread.sleep(5000);
            }catch (InterruptedException e){
                e.printStackTrace();
            }
            //获取各个pod的日志，并截取最后5条
            for(PodInfo podInfo : targetPodInfoList) {
                if(isRequestComplete == true){
                    break;
                }
                System.out.println("[=====] We are now checking POD-LOG:" + podInfo.getName());
                String podLog = getPodLog(podInfo.getName(),"istio-proxy",cluster);
                String[] logsFormatted = podLog.split("\n");
                ArrayList<String> arrayList = new ArrayList<>(Arrays.asList(logsFormatted));
                ArrayList<String> lastSeveralLogs = new ArrayList<>(arrayList.subList(arrayList.size() - 10,arrayList.size()));
                System.out.println("[=====]读取到的日志数量为：" + arrayList.size());
                System.out.println("[=====]截取的日志数量为：" + lastSeveralLogs.size());
                for(String logStr : lastSeveralLogs) {
                    System.out.println("[=======]Log Line -:" + logStr);
                    //并检查日志     response-code 200 && svcName && 接口名称
                    //检查log是否合规
                    if(checkLogCanComfirmRequestComplete(logStr,svcName)){
                        isRequestComplete = true;
                        break;
                    }else{
                        isRequestComplete = false;
                    }
                }
            }
            count += 1;
            if(count > 15){
                isRequestComplete = true;
                System.out.println("没找到这个请求，循环放弃，释放下一个请求");
                count = 0;
            }
        }
        return isRequestComplete;
    }

    private boolean checkLogCanComfirmRequestComplete(String log, String svcName){
        //[2018-04-04T06:18:18.275Z]
        // "GET /helloRestServiceSubOne HTTP/1.1"
        // 200 - 0 29 254 243 "-"
        // "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36"
        // "4939e5f4-6760-9181-afe7-b25e54e57c5f"
        // "rest-service-sub-1:16101"
        // "127.0.0.1:16101"
        //注意：这个判断是有漏洞的
        if(log.contains("200") &&
                log.contains(svcName)){
            return true;
        }else{
            return false;
        }
    }


    //Set the required number of service replicas
    @Override
    public SetServiceReplicasResponse setServiceReplica(SetServiceReplicasRequest setServiceReplicasRequest) {
        Cluster cluster = getClusterByName(setServiceReplicasRequest.getClusterName());
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        SetServiceReplicasResponse response = new SetServiceReplicasResponse();
        List<String> serviceNames = new ArrayList<>();
        //Set the desired number of service replicas
        if(setServiceReplicasRequest.getServiceReplicasSettings() != null){
            for(ServiceReplicasSetting setting : setServiceReplicasRequest.getServiceReplicasSettings()){
                serviceNames.add(setting.getServiceName());
                String apiUrl = String.format("%s/apis/extensions/v1beta1/namespaces/%s/deployments/%s/scale",cluster.getApiServer() ,NAMESPACE,setting.getServiceName());
                System.out.println(String.format("The constructed api url is %s", apiUrl));
                String data ="'[{ \"op\": \"replace\", \"path\": \"/spec/replicas\", \"value\":" +  setting.getNumOfReplicas() + " }]'";

                String[] cmds ={
                        "/bin/sh","-c",String.format("curl -X PATCH -d%s -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure",data,apiUrl,cluster.getToken())
                };
                ProcessBuilder pb = new ProcessBuilder(cmds);
                pb.redirectErrorStream(true);
                Process p;
                try {
                    p = pb.start();
                    BufferedReader br = null;
                    String line = null;

                    //Read the response
                    br = new BufferedReader(new InputStreamReader(p.getInputStream()));
                    StringBuilder responseBuilder = new StringBuilder();
                    boolean record = false;
                    while((line = br.readLine()) != null){
                        if(line.contains("{"))
                            record = true;
                        if(record)
                            responseBuilder.append(line);
                    }
                    //Parse the response to the SetServicesReplicasResponseFromAPI Bean
//                System.out.println(responseBuilder.toString());
                    SetServicesReplicasResponseFromAPI result = JSON.parseObject(responseBuilder.toString(), SetServicesReplicasResponseFromAPI.class);
                    System.out.println(String.format("The kind of the result for service %s is %s",setting.getServiceName(),result.getKind()));
                    br.close();
                } catch (IOException e) {
                    response.setStatus(false);
                    response.setMessage(String.format("Exception: %s", e.getStackTrace()));
                    e.printStackTrace();
                }
            }
        }
        else{
            System.out.println("Error! Please check the parameter in the request!");
        }
        //Check if all the required replicas are ready: running status
        int count = 10;
        boolean b = isAllReady(setServiceReplicasRequest, cluster);
        while(!b && count > 0){
            try{
                //Check every 10 seconds
                Thread.sleep(10000);
            }catch(Exception e){
                e.printStackTrace();
            }
            b = isAllReady(setServiceReplicasRequest, cluster);
            count--;
        }
        if(!b){
            System.out.println("[Set service replicas].There are still some services not ready.");
            response.setMessage("There are still some services not ready.");
            response.setStatus(false);
            return response;
        }
        //Check if all the pods are able to serve
        boolean result = isAllAbleToServe(serviceNames,cluster);
        if(result){
            System.out.println("All the services are able to serve");
        }else{
            System.out.println("There are still some services not able to serve");
        }
        response.setStatus(true);
        response.setMessage("All the required service replicas have been already set!");
        return response;
    }

    //Get all of the services name
    @Override
    public GetServicesListResponse getServicesList(String clusterName) {
        Cluster cluster = getClusterByName(clusterName);
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        GetServicesListResponse response = new GetServicesListResponse();
        //Get the current deployments information
        QueryDeploymentsListResponse deploymentsList = getDeploymentList(NAMESPACE,cluster);
        //Iterate the list and return the result
        List<ServiceWithReplicas> services = new ArrayList<ServiceWithReplicas>();
        if(deploymentsList.getItems() != null && deploymentsList.getItems().size() > 0){
            for(SingleDeploymentInfo singleDeploymentInfo : deploymentsList.getItems()){
                ServiceWithReplicas serviceWithReplicas = new ServiceWithReplicas();
                serviceWithReplicas.setServiceName(singleDeploymentInfo.getMetadata().getName());
                serviceWithReplicas.setNumOfReplicas(singleDeploymentInfo.getStatus().getReadyReplicas());
                services.add(serviceWithReplicas);
            }
        }
        else{
            System.out.println(String.format("There is no deployments in [%s] now!", cluster.getName()));
        }

        System.out.println(String.format("The size of current service is %d",services.size()));
        if(deploymentsList.getItems().size() != 0){
            response.setServices(services);
            response.setMessage("Get the service list successfully!");
            response.setStatus(true);
        }
        else{
            response.setStatus(false);
            response.setMessage("Fail to get the service list!");
        }
        return response;
    }

    //Get the replicas num of the specific services
    @Override
    public GetServiceReplicasResponse getServicesReplicas(GetServiceReplicasRequest getServiceReplicasRequest) {
        Cluster cluster = getClusterByName(getServiceReplicasRequest.getClusterName());
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        GetServiceReplicasResponse response = new GetServiceReplicasResponse();
        //Get the current deployments information
        QueryDeploymentsListResponse deploymentsList = getDeploymentList(NAMESPACE,cluster);
        //Iterate the list and return the result
        List<ServiceWithReplicas> services = new ArrayList<ServiceWithReplicas>();
        if(deploymentsList.getItems() != null && deploymentsList.getItems().size() > 0){
            for(SingleDeploymentInfo singleDeploymentInfo : deploymentsList.getItems()){
                for(String serviceName : getServiceReplicasRequest.getServices()){
                    if(singleDeploymentInfo.getMetadata().getName().equals(serviceName)){
                        ServiceWithReplicas serviceWithReplicas = new ServiceWithReplicas();
                        serviceWithReplicas.setServiceName(serviceName);
                        serviceWithReplicas.setNumOfReplicas(singleDeploymentInfo.getStatus().getReadyReplicas());
                        services.add(serviceWithReplicas);
                        break;
                    }
                }
            }
        }
        else{
            System.out.println(String.format("There is no deployments in [%s] now!", cluster.getName()));
        }

        System.out.println(String.format("The size of current service is %d",services.size()));
        if(services.size() != 0){
            response.setServices(services);
            response.setMessage("Get the replicas of the corresponding services successfully!");
            response.setStatus(true);
        }
        else{
            response.setStatus(false);
            response.setMessage("Fail to get the replicas of the corresponding services!");
        }
        return response;
    }

    //Reserve the services included in the list and delete the others
    @Override
    public ReserveServiceByListResponse reserveServiceByList(ReserveServiceRequest reserveServiceRequest) {
        Cluster cluster = getClusterByName(reserveServiceRequest.getClusterName());
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        ReserveServiceByListResponse response = new ReserveServiceByListResponse();
        response.setStatus(true);
        response.setMessage("Succeed to delete all of the services not contained in the list");
        //Get the current deployments information
        QueryDeploymentsListResponse deploymentsList = getDeploymentList(NAMESPACE, cluster);

        if(deploymentsList.getItems() != null && deploymentsList.getItems().size() > 0){
            for(SingleDeploymentInfo singleDeploymentInfo : deploymentsList.getItems()){
                //Delete the services not contained in the list
                String deploymentName = singleDeploymentInfo.getMetadata().getName();
                if(isDeleted(deploymentName,reserveServiceRequest.getServices())){
                    System.out.println(String.format("The service %s isn't contained in the reserved list. To be deleted",deploymentName ));
                    //Delete the service first
                    deleteService(deploymentName,cluster);
                    //Delete the corresponding pod by set the number of replica to 0
                    boolean result = setServiceReplica(deploymentName, 0,cluster);
                    if(!result){
                        response.setStatus(false);
                        response.setMessage(String.format("Fail to delete the service %s", deploymentName));
                        break;
                    }
                }else{
                    System.out.println(String.format("The service %s is contained in the reserved list. Reserve",deploymentName ));
                }
            }
        }
        else{
            System.out.println(String.format("There is no deployments in [%s] now!", cluster.getName()));
        }
        return response;
    }

    //Set the system to run on single node
    @Override
    public SetRunOnSingleNodeResponse setRunOnSingleNode(String clusterName) {
        Cluster cluster = getClusterByName(clusterName);
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        SetRunOnSingleNodeResponse response = new SetRunOnSingleNodeResponse();
        //Set the default information
        response.setStatus(false);
        response.setMessage("There is no message now!");

        V1NodeList nodeList = getNodeList(cluster);
        List<V1Node> workingNodeList = new ArrayList<V1Node>();

        //Construct the working node list
        if(nodeList.getItems() != null && nodeList.getItems().size() > 0){
            for(V1Node node : nodeList.getItems()){
                System.out.println(String.format("The node name is %s and the role is %s",node.getMetadata().getName(),node.getSpec().getTaints() == null?"Minion":"Master"));
                if(node.getSpec().getTaints() == null)
                    workingNodeList.add(node);
            }
        }else{
            System.out.println(String.format("Fail to get the node info about [%s]", cluster.getName()));
        }


        //Delete the other working nodes and reserve only one and return until all the services are available again
        if(workingNodeList.size() <= 1){
            System.out.println("There is at most one working node. Nothing to do.");
        }else{
            //Delete node
            for(int i = 1; i < workingNodeList.size(); i++){
                V1Node node = workingNodeList.get(i);
                System.out.println(String.format("The node %s is to be deleted",node.getMetadata().getName()));
                deleteNode(node.getMetadata().getName(),cluster);
            }
            //Sleep to let the kubernetes has time to refresh the status
            //TODO: time value to be determined. It seems that 10s is too short。 30s is enough
            try{
                Thread.sleep(30000);
            }catch(Exception e){
                e.printStackTrace();
            }
            //Check whether all of the services are available again in certain internals
            int count = 10;
            boolean b = isAllReady(NAMESPACE,cluster);
            while(!b && count > 0){
                try{
                    //Check every 10 seconds
                    Thread.sleep(10000);
                }catch(Exception e){
                    e.printStackTrace();
                }
                b = isAllReady(NAMESPACE,cluster);
                count--;
            }
            if(!b){
                System.out.println("[Set run on single node].There are still some services not ready.");
                response.setMessage("There are still some services not ready.");
                response.setStatus(false);
                return response;
            }
            //Check if all the service are able to serve
            boolean result = isAllServiceReadyToServe(NAMESPACE, cluster);
            if(result){
                System.out.println("All the services are able to serve");
            }else{
                System.out.println("There are still some services not able to serve");
            }

            response.setStatus(true);
            response.setMessage("The system are now run on single node");
        }
        return response;
    }

    //Get the node list
    @Override
    public GetNodesListResponse getNodesList(String clusterName) {
        Cluster cluster = getClusterByName(clusterName);
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        GetNodesListResponse response = new GetNodesListResponse();
        V1NodeList nodeList = getNodeList(cluster);
        System.out.println(String.format("There are now %d nodes in the cluster now", nodeList.getItems().size()));
        if(nodeList.getItems().size() < 1){
            response.setStatus(false);
            response.setMessage("There is no nodes in the cluster!");
            response.setNodes(null);
        }
        //Construct the nodeinfo list
        List<NodeInfo> nodeInfos = new ArrayList<NodeInfo>();
        for(V1Node node : nodeList.getItems()){
            NodeInfo nodeInfo = new NodeInfo();
            System.out.println(String.format("The node name is %s and the role is %s",node.getMetadata().getName(),node.getSpec().getTaints() == null?"Minion":"Master"));
            //Set the role
            if(node.getSpec().getTaints() != null)
                nodeInfo.setRole("Master");
            else
                nodeInfo.setRole("Minion");
            //Set the name
            nodeInfo.setName(node.getMetadata().getName());
            V1NodeStatus status = node.getStatus();
            //Set the ip
            List<V1NodeAddress> addresses = status.getAddresses();
            for(V1NodeAddress address : addresses){
                if(address.getType().equals("InternalIP")){
                    nodeInfo.setIp(address.getAddress());
                    break;
                }
            }
            //Set the status
            List<V1NodeCondition> conditions = status.getConditions();
            for(V1NodeCondition condition : conditions){
                if(condition.getType().equals("Ready")){
                    if(condition.getStatus().equals("True")){
                        nodeInfo.setStatus("Ready");
                    }
                    else{
                        nodeInfo.setStatus("NotReady");
                    }
                    break;
                }
            }
            //Set the node info
            V1NodeSystemInfo systemInfo = status.getNodeInfo();
            if(systemInfo != null){
                nodeInfo.setContainerRuntimeVersion(systemInfo.getContainerRuntimeVersion());
                nodeInfo.setKubeletVersion(systemInfo.getKubeletVersion());
                nodeInfo.setKubeProxyVersion(systemInfo.getKubeProxyVersion());
                nodeInfo.setOperatingSystem(systemInfo.getOperatingSystem());
                nodeInfo.setOsImage(systemInfo.getOsImage());
            }
            nodeInfos.add(nodeInfo);
        }
        response.setStatus(true);
        response.setMessage("Succeed to get the node list info!");
        response.setNodes(nodeInfos);
        return response;
    }

    //Delete the nodes contained in the list
    @Override
    public DeltaNodeByListResponse deleteNodeByList(DeltaNodeRequest deltaNodeRequest) {
        Cluster cluster = getClusterByName(deltaNodeRequest.getClusterName());
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        DeltaNodeByListResponse response = new DeltaNodeByListResponse();
        List<String> nodeNames = deltaNodeRequest.getNodeNames();
        boolean isSuccess =true;
        for(String nodeName : nodeNames){
            if(!deleteNode(nodeName,cluster))
                isSuccess = false;
        }
        if(isSuccess){
            response.setStatus(true);
            response.setMessage("Succeed to delete all of the nodes in the list!");
        }else{
            response.setStatus(false);
            response.setMessage("Fail to delete some of the nodes in the list!");
        }
        //Sleep to let the kubernetes has time to refresh the status
        //TODO: time value to be determined. It seems that 10s is too short。 30s is enough
        try{
            Thread.sleep(30000);
        }catch(Exception e){
            e.printStackTrace();
        }
        //Check whether all of the services are available again in certain internals
//        while(!isAllReady(NAMESPACE,cluster)){
//            try{
//                //Check every 10 seconds
//                Thread.sleep(10000);
//            }catch(Exception e){
//                e.printStackTrace();
//            }
//        }

        //Check if all the service are able to serve
//        while(!isAllServiceReadyToServe(NAMESPACE,cluster)){
//            try{
//                //Check every 10 seconds
//                Thread.sleep(10000);
//            }catch(Exception e){
//                e.printStackTrace();
//            }
//        }
        //Check whether all of the services are available again in certain internals
        int count = 10;
        boolean b = isAllReady(NAMESPACE,cluster);
        while(!b && count > 0){
            try{
                //Check every 10 seconds
                Thread.sleep(10000);
            }catch(Exception e){
                e.printStackTrace();
            }
            b = isAllReady(NAMESPACE,cluster);
            count--;
        }
        if(!b){
            System.out.println("[Delete node by list].There are still some services not ready.");
            response.setMessage("There are still some services not ready.");
            response.setStatus(false);
            return response;
        }
        //Check if all the service are able to serve
        boolean result = isAllServiceReadyToServe(NAMESPACE, cluster);
        if(result){
            System.out.println("All the services are able to serve");
        }else{
            System.out.println("There are still some services not able to serve");
        }

        return response;
    }

    //Reserve the nodes contained in the list
    @Override
    public DeltaNodeByListResponse reserveNodeByList(DeltaNodeRequest deltaNodeRequest) {
        Cluster cluster = getClusterByName(deltaNodeRequest.getClusterName());
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        DeltaNodeByListResponse response = new DeltaNodeByListResponse();
        List<String> nodeNames = deltaNodeRequest.getNodeNames();
        V1NodeList nodeList = getNodeList(cluster);
        boolean isSuccess =true;
        for(V1Node node : nodeList.getItems()){
            if(node.getSpec().getTaints() != null){
                System.out.println("The master can't be deleted!");
                continue;
            }
            String nodeName = node.getMetadata().getName();
            if(!isExistInNodeList(nodeName,nodeNames)){
                if(!deleteNode(nodeName, cluster))
                    isSuccess = false;
            }
        }
        if(isSuccess){
            response.setStatus(true);
            response.setMessage("Succeed to delete all of the nodes not contained in the list!");
        }else{
            response.setStatus(false);
            response.setMessage("Fail to delete some of the nodes not contained in the list!");
        }
        //Sleep to let the kubernetes has time to refresh the status
        //TODO: time value to be determined. It seems that 10s is too short。 30s is enough
        try{
            Thread.sleep(30000);
        }catch(Exception e){
            e.printStackTrace();
        }
        //Check whether all of the services are available again in certain internals
//        while(!isAllReady(NAMESPACE, cluster)){
//            try{
//                //Check every 10 seconds
//                Thread.sleep(10000);
//            }catch(Exception e){
//                e.printStackTrace();
//            }
//        }

        //Check if all the service are able to serve
//        while(!isAllServiceReadyToServe(NAMESPACE, cluster)){
//            try{
//                //Check every 10 seconds
//                Thread.sleep(10000);
//            }catch(Exception e){
//                e.printStackTrace();
//            }
//        }
//Check whether all of the services are available again in certain internals
        int count = 10;
        boolean b = isAllReady(NAMESPACE,cluster);
        while(!b && count > 0){
            try{
                //Check every 10 seconds
                Thread.sleep(10000);
            }catch(Exception e){
                e.printStackTrace();
            }
            b = isAllReady(NAMESPACE,cluster);
            count--;
        }
        if(!b){
            System.out.println("[Reserve node by list].There are still some services not ready.");
            response.setMessage("There are still some services not ready.");
            response.setStatus(false);
            return response;
        }
        //Check if all the service are able to serve
        boolean result = isAllServiceReadyToServe(NAMESPACE, cluster);
        if(result){
            System.out.println("All the services are able to serve");
        }else{
            System.out.println("There are still some services not able to serve");
        }

        return response;
    }

    //Get the pods info list
    @Override
    public GetPodsListResponse getPodsListAPI(String clusterName) {
        Cluster cluster = getClusterByName(clusterName);
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        GetPodsListResponse response = new GetPodsListResponse();
        V1PodList podList = getPodList("",cluster);
        System.out.println(String.format("There are now %d pods in the cluster now", podList.getItems().size()));
        if(podList.getItems().size() < 1){
            response.setStatus(true);
            response.setMessage("No resource found!");
            response.setPods(null);
        }
        //Construct the podinfo list
        List<PodInfo> podInfos = new ArrayList<PodInfo>();
        for(V1Pod pod : podList.getItems()){
            PodInfo podInfo = new PodInfo();
            podInfo.setName(pod.getMetadata().getName());
            podInfo.setStatus(pod.getStatus().getPhase());
            podInfo.setNodeName(pod.getSpec().getNodeName());
            podInfo.setNodeIP(pod.getStatus().getHostIP());
            podInfo.setPodIP(pod.getStatus().getPodIP());
            podInfo.setStartTime(pod.getStatus().getStartTime());
            podInfos.add(podInfo);
        }
        response.setStatus(true);
        response.setMessage("Successfully get the pod info list!");
        response.setPods(podInfos);
        return response;
    }

    public GetPodsListResponse getPodsList(String namespace, Cluster cluster) {
        GetPodsListResponse response = new GetPodsListResponse();
        V1PodList podList = getPodList(namespace,cluster);
        System.out.println(String.format("There are now %d pods in the cluster now", podList.getItems().size()));
        if(podList.getItems().size() < 1){
            response.setStatus(true);
            response.setMessage("No resource found!");
            response.setPods(null);
        }
        //Construct the podinfo list
        List<PodInfo> podInfos = new ArrayList<PodInfo>();
        for(V1Pod pod : podList.getItems()){
            PodInfo podInfo = new PodInfo();
            podInfo.setName(pod.getMetadata().getName());
            podInfo.setStatus(pod.getStatus().getPhase());
            podInfo.setNodeName(pod.getSpec().getNodeName());
            podInfo.setNodeIP(pod.getStatus().getHostIP());
            podInfo.setPodIP(pod.getStatus().getPodIP());
            podInfo.setStartTime(pod.getStatus().getStartTime());
            podInfos.add(podInfo);
        }
        response.setStatus(true);
        response.setMessage("Successfully get the pod info list!");
        response.setPods(podInfos);
        return response;
    }

    //Get the logs of all pods
    @Override
    public GetPodsLogResponse getPodsLog(String clusterName) {
        Cluster cluster = getClusterByName(clusterName);
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        GetPodsLogResponse response = new GetPodsLogResponse();
        V1PodList podList = getPodList("",cluster);
        System.out.println(String.format("There are now %d pods in the cluster now", podList.getItems().size()));
        if(podList.getItems().size() < 1){
            response.setStatus(true);
            response.setMessage("No resource found!");
            response.setPodLogs(null);
        }
        //Construct the pods log info
        List<PodLog> podLogs = new ArrayList<PodLog>();
        List<V1Container> containers;
        if(podList.getItems() != null && podList.getItems().size() > 0){
            for(V1Pod pod : podList.getItems()){
                PodLog podLog = new PodLog();
                String podName = pod.getMetadata().getName();
                String containerName = "";
                containers = pod.getSpec().getContainers();
                if(containers.size() > 0){
                    for(V1Container container : containers){
                        if(!container.getName().equals("istio-proxy")){
                            containerName = container.getName();
                            break;
                        }
                    }
                }
                podLog.setPodName(podName);
                String logs = getPodLog(podName,containerName,cluster);
                podLog.setLogs(logs);
                podLogs.add(podLog);
            }
        }else{
            System.out.println(String.format("There is no pod in the [%s]", cluster.getName()));
        }

        response.setStatus(true);
        response.setMessage("Successfully to get the pods log info!");
        response.setPodLogs(podLogs);
        return response;
    }

    //Get the log of the single pod
    @Override
    public GetSinglePodLogResponse getSinglePodLog(GetSinglePodLogRequest getSinglePodLogRequest) {
        Cluster cluster = getClusterByName(getSinglePodLogRequest.getClusterName());
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        GetSinglePodLogResponse response = new GetSinglePodLogResponse();
        response.setStatus(false);
        response.setMessage("Fail to get the corresponding pod's log!");
        response.setPodLog(null);

        //Get the pod name and the container name
        PodLog podLog = new PodLog();
        String podName = getSinglePodLogRequest.getPodName();
        String containerName = "";
        V1Pod pod = getPodInfo(podName,cluster);
        if(pod != null){
            List<V1Container> containers = pod.getSpec().getContainers();
            if(containers.size() > 0){
                for(V1Container container : containers){
                    if(!container.getName().equals("istio-proxy")){
                        containerName = container.getName();
                        break;
                    }
                }
            }else{
                response.setStatus(false);
                response.setMessage("There are no containers in the specified pod!");
                response.setPodLog(null);
                return response;
            }

            //Get the log with the pod name and container name
            String log = getPodLog(podName,containerName,cluster);
            if(!log.equals("")){
                podLog.setPodName(getSinglePodLogRequest.getPodName());
                podLog.setLogs(log);
                response.setStatus(true);
                response.setMessage("Successfully get the corresponding pod's log!");
                response.setPodLog(podLog);
            }
        }else{
            response.setStatus(false);
            response.setMessage("There is no such pod in the cluster!");
            response.setPodLog(null);
        }
        return response;
    }

    //Restart the service: current zipkin only
    @Override
    public RestartServiceResponse restartService(String clusterName) {
        Cluster cluster = getClusterByName(clusterName);
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        boolean isExist = false;
        RestartServiceResponse response = new RestartServiceResponse();
        V1PodList podList = getPodList("istio-system",cluster);
        //Delete the zipkin pod
        if(podList.getItems() != null && podList.getItems().size() > 0){
            for(V1Pod pod : podList.getItems()) {
                String podName = pod.getMetadata().getName();
                if(podName.contains("zipkin")){
                    isExist = true;
                    //Delete the pod corresponding to the service
                    boolean deleteResult = deletePod("istio-system",podName,cluster);
                    if(!deleteResult){
                        response.setStatus(false);
                        response.setMessage("Fail to restart zipkin!");
                        return response;
                    }
                    break;
                }
            }
        }

        if(isExist){
            //Wait for the zipkin restart
            try{
                Thread.sleep(6000);
            }catch(Exception e){
                e.printStackTrace();
            }
            //Check whether all of the services are available again in certain internals
            while(!isAllReady("istio-system",cluster)){
                try{
                    //Check every 10 seconds
                    Thread.sleep(3000);
                }catch(Exception e){
                    e.printStackTrace();
                }
            }
            response.setStatus(true);
            response.setMessage("The zipkin has been restarted successfully!");
        }else{
            response.setStatus(false);
            response.setMessage("The zipkin service doesn't exist!");
        }
        return response;
    }

    //Get the service list and the config settings
    @Override
    public GetServicesAndConfigResponse getServicesAndConfig(String clusterName) {
        Cluster cluster = getClusterByName(clusterName);
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        GetServicesAndConfigResponse response = new GetServicesAndConfigResponse();
        //Get the current deployments information
        QueryDeploymentsListResponse deploymentsList = getDeploymentList(NAMESPACE,cluster);
        //Iterate the list and return the result
        List<ServiceWithConfig> services = new ArrayList<ServiceWithConfig>();
        if(deploymentsList.getItems() != null && deploymentsList.getItems().size() > 0){
            for(SingleDeploymentInfo singleDeploymentInfo : deploymentsList.getItems()){
                V1ResourceRequirements resourceRequirements = singleDeploymentInfo.getSpec().getTemplate().getSpec().getContainers().get(0).getResources();
                ServiceWithConfig serviceWithConfig = new ServiceWithConfig();
                serviceWithConfig.setServiceName(singleDeploymentInfo.getMetadata().getName());
                serviceWithConfig.setLimits(resourceRequirements.getLimits());
                serviceWithConfig.setRequests(resourceRequirements.getRequests());
                services.add(serviceWithConfig);
            }
        }else{
            System.out.println(String.format("There is no deployments in [%s] now!", cluster.getName()));
        }

        System.out.println(String.format("The size of current service is %d",services.size()));
        if(services.size() != 0){
            response.setServices(services);
            response.setMessage("Get the services and the corresponding config successfully!");
            response.setStatus(true);
        }
        else{
            response.setStatus(false);
            response.setMessage("Fail to get the services and the corresponding config!");
        }
        return response;
    }

    //Delta the cpu and memory resource
    @Override
    public DeltaCMResourceResponse deltaCMResource(DeltaCMResourceRequest deltaCMResourceRequest) {
        Cluster cluster = getClusterByName(deltaCMResourceRequest.getClusterName());
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        DeltaCMResourceResponse response = new DeltaCMResourceResponse();
        //Get the current deployments information
        QueryDeploymentsListResponse deploymentsList = getDeploymentList(NAMESPACE,cluster);

        List<String> serviceNames = new ArrayList<>();
        //Check if the resource setting exists
        if(deploymentsList.getItems() != null && deploymentsList.getItems().size() > 0 && deltaCMResourceRequest.getDeltaRequests() != null){
            for(NewSingleDeltaCMResourceRequest request : deltaCMResourceRequest.getDeltaRequests()){
                serviceNames.add(request.getServiceName());
                for(SingleDeploymentInfo singleDeploymentInfo : deploymentsList.getItems()){
                    if(singleDeploymentInfo.getMetadata().getName().equals(request.getServiceName())){
                        V1ResourceRequirements resourceRequirements = singleDeploymentInfo.getSpec().getTemplate().getSpec().getContainers().get(0).getResources();
                        if(resourceRequirements.getLimits() == null || resourceRequirements.getRequests() == null){
                            response.setMessage("There is no corresponding config in the cluster!");
                            response.setStatus(false);
                            //TODO: Add the config through add option
                        }
                        else{
                            boolean result = deltaCMResource(NAMESPACE,request, cluster);
                            if(result){
                                response.setStatus(true);
                                response.setMessage("The config has been deltaed successfully!");
                            }
                        }
                    }
                }
            }
        }

        //Wait for the pod restart
        try{
            Thread.sleep(1000);
        }catch(Exception e){
            e.printStackTrace();
        }
        //Check whether all of the services are available again in certain internals
        int count = 10;
        boolean b = isAllReady(NAMESPACE,cluster);
        while(!b && count > 0){
            try{
                //Check every 2 seconds
                Thread.sleep(2000);
            }catch(Exception e){
                e.printStackTrace();
            }
            b = isAllReady(NAMESPACE,cluster);
            count--;
        }
        if(!b){
            System.out.println("There are still some pods are not ready after config delta");
            response.setMessage("There are still some pods are not ready after config delta");
            response.setStatus(false);
            return response;
        }
        //Check if all the pods are able to serve
        boolean result = isAllAbleToServe(serviceNames,cluster);
        if(result){
            System.out.println("All the services are able to serve");
        }else{
            System.out.println("There are still some services not able to serve");
        }

        return response;
    }

    //Get the service with endpoints
    @Override
    public ServiceWithEndpointsResponse getServiceWithEndpoints(String clusterName) {
        Cluster cluster = getClusterByName(clusterName);
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        ServiceWithEndpointsResponse response = new ServiceWithEndpointsResponse();
        //Get the current endpoints list
        V1EndpointsList endpointsList = getEndpointsList(NAMESPACE,cluster);
        //Iterate the list and return the result
        List<ServiceWithEndpoints> services = new ArrayList<ServiceWithEndpoints>();
        if(endpointsList.getItems() != null && endpointsList.getItems().size() > 0){
            for(V1Endpoints endpoints : endpointsList.getItems()){
                ServiceWithEndpoints serviceWithEndpoints = new ServiceWithEndpoints();
                //Set service name
                serviceWithEndpoints.setServiceName(endpoints.getMetadata().getName());
                //Set service endpoints
                List<String> endpointsListOfService = new ArrayList<>();
                if(endpoints.getSubsets().size() > 0){
                    V1EndpointSubset subset = endpoints.getSubsets().get(0);
                    if(subset.getAddresses() != null && subset.getPorts() != null){
                        for(V1EndpointAddress v1EndpointAddress : subset.getAddresses()){
                            String ip = v1EndpointAddress.getIp();
                            for(V1EndpointPort v1EndpointPort : subset.getPorts()){
                                endpointsListOfService.add(ip + ":" + v1EndpointPort.getPort());
                            }
                        }
                        serviceWithEndpoints.setEndPoints(endpointsListOfService);
                    }
                }
                services.add(serviceWithEndpoints);
            }
        }

        System.out.println(String.format("The size of current service is %d",services.size()));
        if(services.size() != 0){
            response.setServices(services);
            response.setMessage("Get the services and the corresponding endpoints successfully!");
            response.setStatus(true);
        }
        else{
            response.setStatus(false);
            response.setMessage("Fail to get the services and the corresponding endpoints!");
        }
        return response;
    }

    @Override
    public ServiceWithEndpointsResponse getSpecificServiceWithEndpoints(ReserveServiceRequest reserveServiceRequest) {
        Cluster cluster = getClusterByName(reserveServiceRequest.getClusterName());
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        ServiceWithEndpointsResponse response = new ServiceWithEndpointsResponse();
        List<ServiceWithEndpoints> services = new ArrayList<ServiceWithEndpoints>();
        for(String serviceName : reserveServiceRequest.getServices()){
            V1Endpoints endpoints = getSingleServiceEndpoints(NAMESPACE,serviceName,cluster);
            ServiceWithEndpoints serviceWithEndpoints = new ServiceWithEndpoints();
            //Set service name
            serviceWithEndpoints.setServiceName(endpoints.getMetadata().getName());
            //Set service endpoints
            List<String> endpointsListOfService = new ArrayList<>();
            if(endpoints.getSubsets().size() > 0){
                V1EndpointSubset subset = endpoints.getSubsets().get(0);
                if(subset.getAddresses() != null && subset.getPorts() != null){
                    for(V1EndpointAddress v1EndpointAddress : subset.getAddresses()){
                        String ip = v1EndpointAddress.getIp();
                        for(V1EndpointPort v1EndpointPort : subset.getPorts()){
                            endpointsListOfService.add(ip + ":" + v1EndpointPort.getPort());
                        }
                    }
                }
                serviceWithEndpoints.setEndPoints(endpointsListOfService);
            }
            services.add(serviceWithEndpoints);
        }
        response.setStatus(true);
        response.setMessage("Successfully to get the endpoints list of specific services");
        response.setServices(services);
        return response;
    }

    //Delta the config and replicas of service at the same time
    @Override
    public SimpleResponse deltaAll(DeltaAllRequest deltaAllRequest) {
        Cluster cluster = getClusterByName(deltaAllRequest.getClusterName());
        System.out.println(String.format("The cluster to operate is: %s", cluster.getName()));
        SimpleResponse response = new SimpleResponse();
        //Get the current deployments information
        QueryDeploymentsListResponse deploymentsList = getDeploymentList(NAMESPACE,cluster);

        List<String> serviceNames = new ArrayList<>();
        //Check if the resource setting exists
        if(deploymentsList.getItems() != null && deploymentsList.getItems().size() > 0 && deltaAllRequest.getDeltaRequests() != null){
            for(SingleDeltaAllRequest request : deltaAllRequest.getDeltaRequests()){
                serviceNames.add(request.getServiceName());
                for(SingleDeploymentInfo singleDeploymentInfo : deploymentsList.getItems()){
                    if(singleDeploymentInfo.getMetadata().getName().equals(request.getServiceName())){
                        boolean result = deltaAllProcess(NAMESPACE,request, cluster);
                        if(result){
                            response.setStatus(true);
                            response.setMessage("The config has been deltaed successfully!");
                        }
                    }
                }
            }
        }

        //Wait for the pod restart
        try{
            Thread.sleep(1000);
        }catch(Exception e){
            e.printStackTrace();
        }
        //Check whether all of the services are available again in certain internals
        int count = 10;
        boolean b = isAllReady(NAMESPACE,cluster);
        while(!b && count > 0){
            try{
                //Check every 2 seconds
                Thread.sleep(2000);
            }catch(Exception e){
                e.printStackTrace();
            }
            b = isAllReady(NAMESPACE,cluster);
            count--;
        }
        if(!b){
            System.out.println("There are still some pods are not ready after delta all");
            response.setMessage("There are still some pods are not ready after delta all");
            response.setStatus(false);
            return response;
        }
        //Check if all the pods are able to serve
        boolean result = isAllAbleToServe(serviceNames,cluster);
        if(result){
            System.out.println("All the services are able to serve");
        }else{
            System.out.println("There are still some services not able to serve");
        }

        return response;
    }

    //Check if all the services in the list are able to serve
    private boolean isAllAbleToServe(List<String> serviceNames,Cluster cluster){
        System.out.println(String.format("The service list to be checked are %s", serviceNames.toString()));
        boolean symbol, result = true;
        for(String serviceName : serviceNames){
            System.out.println(String.format("The service to be checked is %s", serviceName));
            V1Endpoints endpoints = getSingleServiceEndpoints(NAMESPACE,serviceName,cluster);
            //Get service endpoints
            List<String> endpointsListOfService = new ArrayList<>();
            if(endpoints.getSubsets().size() > 0){
                V1EndpointSubset subset = endpoints.getSubsets().get(0);
                if(subset.getAddresses() != null && subset.getPorts() != null){
                    for(V1EndpointAddress v1EndpointAddress : subset.getAddresses()){
                        String ip = v1EndpointAddress.getIp();
                        for(V1EndpointPort v1EndpointPort : subset.getPorts()){
                            endpointsListOfService.add(ip + ":" + v1EndpointPort.getPort());
                        }
                    }
                }
                else{
                    System.out.println(String.format("The service [%s] doesn't have ready address!",serviceName));
                }
            }
            //Check if all the endpoints in the list are able to serve
            int count = 3;
            symbol = isAllEndpointsAbleToServe(endpointsListOfService,cluster);
            while(!symbol && count > 0){
                try{
                    //Check every 10 seconds
                    Thread.sleep(10000);
                }catch(Exception e){
                    e.printStackTrace();
                }
                count--;
                symbol = isAllEndpointsAbleToServe(endpointsListOfService, cluster);
            }
            if(symbol){
                System.out.println(String.format("All of the endpoints of service:%s are able to serve.", serviceName));
            }else{
                result = false;
                System.out.println(String.format("Timeout! The endpoints of service:%s still have problem.", serviceName));
            }
        }
        return result;
    }

    //Return the cluster info corresponding to the cluster name
    private Cluster getClusterByName(String clusterName){
        if(myConfig.getClusters() != null){
            for(Cluster cluster : myConfig.getClusters()){
                if(cluster.getName().equals(clusterName))
                    return cluster;
            }
        }
        else{
            System.out.println("There is some error in the application.yml. Please check!");
        }
        System.out.println(String.format("The cluster corresponding to the name [%s] doesn't exist! Please check."));
        return null;
    }

    //Check if all the endpoints in the list are able to serve
    private boolean isAllEndpointsAbleToServe(List<String> endpointsList, Cluster cluster){
        boolean result = true;
        RemoteExecuteCommand rec = new RemoteExecuteCommand(cluster.getMasterIp(), cluster.getUsername(),cluster.getPasswd());
        for(String endpoints : endpointsList){
            String command = String.format("curl -X GET %s/welcome", endpoints);
            System.out.println(String.format("The command to check the endpoints is %s", command));
            String executeResult = rec.execute(command);
            System.out.println(String.format("The return string is %s", executeResult));
            if(executeResult != null && !executeResult.equals("") && (executeResult.contains("welcome") || executeResult.contains("Welcome"))){

            }else{
                System.out.println(String.format("The endpoints:%s is not ready to serve", endpoints));
                return false;
            }
        }
        return result;
    }

    //Get the endpoints list of specific service
    private V1Endpoints getSingleServiceEndpoints(String namespace, String serviceName, Cluster cluster){
        V1Endpoints endpoints = new V1Endpoints();
        String filePath = "/app/get_endpoints_list_of_single_service_" + cluster.getName() + System.currentTimeMillis()+ ".json";
        String apiUrl = String.format("%s/api/v1/namespaces/%s/endpoints/%s",cluster.getApiServer(),namespace,serviceName);
        System.out.println(String.format("The constructed api url for getting the endpoints of specific service is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,cluster.getToken(),filePath)
        };
        ProcessBuilder pb = new ProcessBuilder(cmds);
        pb.redirectErrorStream(true);
        Process p;
        try {
            p = pb.start();
            p.waitFor();
            String json = readWholeFile(filePath);
            endpoints = JSON.parseObject(json,V1Endpoints.class);
        } catch (IOException e) {
            e.printStackTrace();
        }catch(InterruptedException e){
            e.printStackTrace();
        }
        return endpoints;
    }

    //Get the endpoints list of all services
    private V1EndpointsList getEndpointsList(String namespace, Cluster cluster){
        V1EndpointsList endpointsList = new V1EndpointsList();
        String filePath = "/app/get_endpoints_list_" + cluster.getName() + System.currentTimeMillis()+ ".json";
        String apiUrl = String.format("%s/api/v1/namespaces/%s/endpoints",cluster.getApiServer(),namespace);
        System.out.println(String.format("The constructed api url for getting the endpoints list is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,cluster.getToken(),filePath)
        };
        ProcessBuilder pb = new ProcessBuilder(cmds);
        pb.redirectErrorStream(true);
        Process p;
        try {
            p = pb.start();
            p.waitFor();
            String json = readWholeFile(filePath);
            endpointsList = JSON.parseObject(json,V1EndpointsList.class);
        } catch (IOException e) {
            e.printStackTrace();
        }catch(InterruptedException e){
            e.printStackTrace();
        }
        return endpointsList;
    }

    //Not Used Any More: Delta the CPU and memory of service
    @Deprecated
    private boolean deltaCMResource(String namespace,SingleDeltaCMResourceRequest request, Cluster cluster){
        boolean isSuccess = true;
        SingleDeploymentInfo result;
        String filePath = "/app/delta_cmconfig_result_" + cluster.getName() + System.currentTimeMillis()+ ".json";
        String apiUrl = String.format("%s/apis/apps/v1beta1/namespaces/%s/deployments/%s",cluster.getApiServer(),namespace, request.getServiceName());
        System.out.println(String.format("The constructed api url for deltaing config is %s", apiUrl));
        String data = String.format("'[{ \"op\": \"replace\", \"path\": \"/spec/template/spec/containers/0/resources/%s/%s\", \"value\": \"%s\"}]'",
                request.getType(),request.getKey(),request.getValue());
        System.out.println(String.format("The constructed data for deltaing config is %s", data));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X PATCH -d%s -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",
                data,apiUrl,cluster.getToken(),filePath)
        };
        ProcessBuilder pb = new ProcessBuilder(cmds);
        pb.redirectErrorStream(true);
        Process p;
        try {
            p = pb.start();
            p.waitFor();

            String json = readWholeFile(filePath);
            //Parse the response to the SetServicesReplicasResponseFromAPI Bean
//            System.out.println(json);
            result = JSON.parseObject(json,SingleDeploymentInfo.class);
        } catch (Exception e) {
            isSuccess = false;
            e.printStackTrace();
        }
        return isSuccess;
    }

    //New: Delta the CPU and memory of service. Restart only once.
    private boolean deltaCMResource(String namespace,NewSingleDeltaCMResourceRequest request, Cluster cluster){
        boolean isSuccess = true;
        SingleDeploymentInfo result;
        String filePath = "/app/delta_cmconfig_result_" + cluster.getName() + System.currentTimeMillis()+ ".json";
        String apiUrl = String.format("%s/apis/apps/v1beta1/namespaces/%s/deployments/%s",cluster.getApiServer(),namespace, request.getServiceName());
        System.out.println(String.format("The constructed api url for deltaing config is %s", apiUrl));
        //Add the type: limits and requests
        List<CMConfig> configs = request.getConfigs();

        String[] cmds = new String[]{};
        //Delta limits and requests at the same time
        if(configs.size() > 1){
            cmds = new String[]{
                    "/bin/sh","-c",String.format("curl -X PATCH -d \"[{\\\"op\\\":\\\"replace\\\",\\\"path\\\":\\\"/spec/template/spec/containers/0/resources\\\",\\\"value\\\":{\\\"%s\\\":{\\\"%s\\\":\\\"%s\\\", \\\"%s\\\":\\\"%s\\\"},\\\"%s\\\":{\\\"%s\\\":\\\"%s\\\", \\\"%s\\\":\\\"%s\\\"}}}]\" -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",
                    configs.get(0).getType(),configs.get(0).getValues().get(0).getKey(),configs.get(0).getValues().get(0).getValue(),configs.get(0).getValues().get(1).getKey(),configs.get(0).getValues().get(1).getValue(),
                    configs.get(1).getType(),configs.get(1).getValues().get(0).getKey(),configs.get(1).getValues().get(0).getValue(),configs.get(1).getValues().get(1).getKey(),configs.get(1).getValues().get(1).getValue(),
                    apiUrl,cluster.getToken(),filePath)
            };
        }
        //Delta limits or requests
        else{
            System.out.println("[Delta All] Delta limits or requests only");
            if(configs.get(0).getValues().size() > 1){
                cmds = new String[]{
                        "/bin/sh","-c",String.format("curl -X PATCH -d \"[" +
                                "{\\\"op\\\":\\\"replace\\\"," +
                                "\\\"path\\\":\\\"/spec/template/spec/containers/0/resources/%s\\\"," +
                                "\\\"value\\\":{\\\"%s\\\":\\\"%s\\\", \\\"%s\\\":\\\"%s\\\"}}" +
                                "]\" -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",
                        configs.get(0).getType(),configs.get(0).getValues().get(0).getKey(),configs.get(0).getValues().get(0).getValue(),configs.get(0).getValues().get(1).getKey(),configs.get(0).getValues().get(1).getValue(),
                        apiUrl,cluster.getToken(),filePath)
                };
            }else if(configs.get(0).getValues().size() == 1){
                //Check if exist limits and memory at the same time
                SingleDeploymentInfo singleDeploymentInfo = getSingleDeployment(namespace, request.getServiceName(),cluster);
                V1ResourceRequirements resourceRequirements = singleDeploymentInfo.getSpec().getTemplate().getSpec().getContainers().get(0).getResources();
                if(resourceRequirements.getLimits().size() == 1 || resourceRequirements.getRequests().size() == 1){
                    //Only one config: cpu or memory
                    cmds = new String[]{
                            "/bin/sh","-c",String.format("curl -X PATCH -d \"[" +
                                    "{\\\"op\\\":\\\"replace\\\"," +
                                    "\\\"path\\\":\\\"/spec/template/spec/containers/0/resources/%s\\\"," +
                                    "\\\"value\\\":{\\\"%s\\\":\\\"%s\\\"}}" +
                                    "]\" -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",
                            configs.get(0).getType(),configs.get(0).getValues().get(0).getKey(),configs.get(0).getValues().get(0).getValue(),
                            apiUrl,cluster.getToken(),filePath)
                    };
                }else{
                    String key, value;
                    if(configs.get(0).getValues().get(0).getKey().equals("memory")){
                        //Add the origin cpu
                        key = "cpu";
                        value = resourceRequirements.getLimits().get("cpu");
                    }else{
                        //Add the origin memory
                        key = "memory";
                        value = resourceRequirements.getLimits().get("memory");
                    }
                    cmds = new String[]{
                            "/bin/sh","-c",String.format("curl -X PATCH -d \"[" +
                                    "{\\\"op\\\":\\\"replace\\\"," +
                                    "\\\"path\\\":\\\"/spec/template/spec/containers/0/resources/%s\\\"," +
                                    "\\\"value\\\":{\\\"%s\\\":\\\"%s\\\", \\\"%s\\\":\\\"%s\\\"}}" +
                                    "]\" -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",
                            configs.get(0).getType(),configs.get(0).getValues().get(0).getKey(),configs.get(0).getValues().get(0).getValue(),key,value,
                            apiUrl,cluster.getToken(),filePath)
                    };
                }
            }
        }
//        String data = String.format("'[{ \"op\": \"replace\", \"path\": \"/spec/template/spec/containers/0/resources/%s/%s\", \"value\": \"%s\"}]'",
//                request.getType(),request.getKey(),request.getValue());
//        System.out.println(String.format("The constructed data for deltaing config is %s", data));
//        cmds ={
//                "/bin/sh","-c",String.format("curl -X PATCH -d%s -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",
//                data,apiUrl,cluster.getToken(),filePath)
//        };
//        System.out.println(String.format("The constructed command for deltaing config is %s", cmds[2]));
        ProcessBuilder pb = new ProcessBuilder(cmds);
        pb.redirectErrorStream(true);
        Process p;
        try {
            p = pb.start();
            p.waitFor();

            String json = readWholeFile(filePath);
            //Parse the response to the SetServicesReplicasResponseFromAPI Bean
//            System.out.println(json);
            result = JSON.parseObject(json,SingleDeploymentInfo.class);
        } catch (Exception e) {
            isSuccess = false;
            e.printStackTrace();
        }
        return isSuccess;
    }


    //Get single deployment info
    private SingleDeploymentInfo getSingleDeployment(String namespace, String serviceName,Cluster cluster){
        SingleDeploymentInfo result = new SingleDeploymentInfo();
        String filePath = "/app/get_single_deployment_result_" + cluster.getName() + System.currentTimeMillis()+ ".json";
        String apiUrl = String.format("%s/apis/apps/v1beta1/namespaces/%s/deployments/%s",cluster.getApiServer(),namespace, serviceName);
        System.out.println(String.format("The constructed api url for deltaing all is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,cluster.getToken(),filePath)
        };
        ProcessBuilder pb = new ProcessBuilder(cmds);
        pb.redirectErrorStream(true);
        Process p;
        try {
            p = pb.start();
            p.waitFor();
            String json = readWholeFile(filePath);
            result = JSON.parseObject(json,SingleDeploymentInfo.class);
        } catch (IOException e) {
            e.printStackTrace();
        }catch(InterruptedException e){
            e.printStackTrace();
        }
        return result;
    }

    //Delta the config and instance of service at the same time
    private boolean deltaAllProcess(String namespace,SingleDeltaAllRequest request, Cluster cluster){
        boolean isSuccess = true;
        SingleDeploymentInfo result;
        String filePath = "/app/delta_all_result_" + cluster.getName() + System.currentTimeMillis()+ ".json";
        String apiUrl = String.format("%s/apis/apps/v1beta1/namespaces/%s/deployments/%s",cluster.getApiServer(),namespace, request.getServiceName());
        System.out.println(String.format("The constructed api url for deltaing all is %s", apiUrl));
        //Add the type: limits and requests
        List<CMConfig> configs = request.getConfigs();
        int num = request.getNumOfReplicas();

        String[] cmds = new String[]{};
        //Delta limits and requests , with instance at the same time
        if(configs == null || configs.size() == 0){
            if(num <= 0){
                cmds = new String[]{};
                System.out.println("There is no need to delta!");
            }
            else{
                //Delta instance only
                System.out.println("[Delta All] Delta instance only");
                cmds = new String[]{
                        "/bin/sh","-c",String.format("curl -X PATCH -d \"[" +
                                "{\\\"op\\\":\\\"replace\\\"," +
                                "\\\"path\\\":\\\"/spec/replicas\\\"," +
                                "\\\"value\\\": %d}" +
                                "]\" -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",
                        request.getNumOfReplicas(), apiUrl,cluster.getToken(),filePath)
                };
            }
        }
        else{
            if(configs.size() == 1){
                //Delta limits or requests, with instance
                if(num > 0){
                    System.out.println("[Delta All] Delta limits or requests, with instance");
                    if(configs.get(0).getValues().size() > 1){
                        cmds = new String[]{
                                "/bin/sh","-c",String.format("curl -X PATCH -d \"[" +
                                        "{\\\"op\\\":\\\"replace\\\"," +
                                        "\\\"path\\\":\\\"/spec/template/spec/containers/0/resources/%s\\\"," +
                                        "\\\"value\\\":{\\\"%s\\\":\\\"%s\\\", \\\"%s\\\":\\\"%s\\\"}}," +
                                        "{\\\"op\\\":\\\"replace\\\"," +
                                        "\\\"path\\\":\\\"/spec/replicas\\\"," +
                                        "\\\"value\\\": %d}" +
                                        "]\" -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",
                                configs.get(0).getType(),configs.get(0).getValues().get(0).getKey(),configs.get(0).getValues().get(0).getValue(),configs.get(0).getValues().get(1).getKey(),configs.get(0).getValues().get(1).getValue(),
                                request.getNumOfReplicas(),
                                apiUrl,cluster.getToken(),filePath)
                        };
                    }else if(configs.get(0).getValues().size() == 1){
                        //Check if exist limits and memory at the same time
                        SingleDeploymentInfo singleDeploymentInfo = getSingleDeployment(namespace, request.getServiceName(),cluster);
                        V1ResourceRequirements resourceRequirements = singleDeploymentInfo.getSpec().getTemplate().getSpec().getContainers().get(0).getResources();
                        if(resourceRequirements.getLimits().size() == 1 || resourceRequirements.getRequests().size() == 1){
                            //Only one config: cpu or memory
                            cmds = new String[]{
                                    "/bin/sh","-c",String.format("curl -X PATCH -d \"[" +
                                            "{\\\"op\\\":\\\"replace\\\"," +
                                            "\\\"path\\\":\\\"/spec/template/spec/containers/0/resources/%s\\\"," +
                                            "\\\"value\\\":{\\\"%s\\\":\\\"%s\\\"}}," +
                                            "{\\\"op\\\":\\\"replace\\\"," +
                                            "\\\"path\\\":\\\"/spec/replicas\\\"," +
                                            "\\\"value\\\": %d}" +
                                            "]\" -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",
                                    configs.get(0).getType(),configs.get(0).getValues().get(0).getKey(),configs.get(0).getValues().get(0).getValue(),
                                    request.getNumOfReplicas(),
                                    apiUrl,cluster.getToken(),filePath)
                            };
                        }else{
                            String key, value;
                            if(configs.get(0).getValues().get(0).getKey().equals("memory")){
                                //Add the origin cpu
                                key = "cpu";
                                value = resourceRequirements.getLimits().get("cpu");
                            }else{
                                //Add the origin memory
                                key = "memory";
                                value = resourceRequirements.getLimits().get("memory");
                            }
                            cmds = new String[]{
                                    "/bin/sh","-c",String.format("curl -X PATCH -d \"[" +
                                            "{\\\"op\\\":\\\"replace\\\"," +
                                            "\\\"path\\\":\\\"/spec/template/spec/containers/0/resources/%s\\\"," +
                                            "\\\"value\\\":{\\\"%s\\\":\\\"%s\\\", \\\"%s\\\":\\\"%s\\\"}}," +
                                            "{\\\"op\\\":\\\"replace\\\"," +
                                            "\\\"path\\\":\\\"/spec/replicas\\\"," +
                                            "\\\"value\\\": %d}" +
                                            "]\" -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",
                                    configs.get(0).getType(),configs.get(0).getValues().get(0).getKey(),configs.get(0).getValues().get(0).getValue(),key,value,
                                    request.getNumOfReplicas(),
                                    apiUrl,cluster.getToken(),filePath)
                            };
                        }
                    }

                }
                //Delta limits or requests only
                else{
                    System.out.println("[Delta All] Delta limits or requests only");
                    if(configs.get(0).getValues().size() > 1){
                        cmds = new String[]{
                                "/bin/sh","-c",String.format("curl -X PATCH -d \"[" +
                                        "{\\\"op\\\":\\\"replace\\\"," +
                                        "\\\"path\\\":\\\"/spec/template/spec/containers/0/resources/%s\\\"," +
                                        "\\\"value\\\":{\\\"%s\\\":\\\"%s\\\", \\\"%s\\\":\\\"%s\\\"}}" +
                                        "]\" -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",
                                configs.get(0).getType(),configs.get(0).getValues().get(0).getKey(),configs.get(0).getValues().get(0).getValue(),configs.get(0).getValues().get(1).getKey(),configs.get(0).getValues().get(1).getValue(),
                                apiUrl,cluster.getToken(),filePath)
                        };
                    }else if(configs.get(0).getValues().size() == 1){
                        //Check if exist limits and memory at the same time
                        SingleDeploymentInfo singleDeploymentInfo = getSingleDeployment(namespace, request.getServiceName(),cluster);
                        V1ResourceRequirements resourceRequirements = singleDeploymentInfo.getSpec().getTemplate().getSpec().getContainers().get(0).getResources();
                        if(resourceRequirements.getLimits().size() == 1 || resourceRequirements.getRequests().size() == 1){
                            //Only one config: cpu or memory
                            cmds = new String[]{
                                    "/bin/sh","-c",String.format("curl -X PATCH -d \"[" +
                                            "{\\\"op\\\":\\\"replace\\\"," +
                                            "\\\"path\\\":\\\"/spec/template/spec/containers/0/resources/%s\\\"," +
                                            "\\\"value\\\":{\\\"%s\\\":\\\"%s\\\"}}" +
                                            "]\" -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",
                                    configs.get(0).getType(),configs.get(0).getValues().get(0).getKey(),configs.get(0).getValues().get(0).getValue(),
                                    apiUrl,cluster.getToken(),filePath)
                            };
                        }else{
                            String key, value;
                            if(configs.get(0).getValues().get(0).getKey().equals("memory")){
                                //Add the origin cpu
                                key = "cpu";
                                value = resourceRequirements.getLimits().get("cpu");
                            }else{
                                //Add the origin memory
                                key = "memory";
                                value = resourceRequirements.getLimits().get("memory");
                            }
                            cmds = new String[]{
                                    "/bin/sh","-c",String.format("curl -X PATCH -d \"[" +
                                            "{\\\"op\\\":\\\"replace\\\"," +
                                            "\\\"path\\\":\\\"/spec/template/spec/containers/0/resources/%s\\\"," +
                                            "\\\"value\\\":{\\\"%s\\\":\\\"%s\\\", \\\"%s\\\":\\\"%s\\\"}}" +
                                            "]\" -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",
                                    configs.get(0).getType(),configs.get(0).getValues().get(0).getKey(),configs.get(0).getValues().get(0).getValue(),key,value,
                                    apiUrl,cluster.getToken(),filePath)
                            };
                        }
                    }

                }
            }
            else{
                //Delta limits and requests, with instance
                if(num > 0){
                    System.out.println("[Delta All] Delta limits and requests, with instance");
                    cmds = new String[]{
                            "/bin/sh","-c",String.format("curl -X PATCH -d \"[" +
                                    "{\\\"op\\\":\\\"replace\\\"," +
                                    "\\\"path\\\":\\\"/spec/template/spec/containers/0/resources\\\"," +
                                    "\\\"value\\\":{\\\"%s\\\":{\\\"%s\\\":\\\"%s\\\", \\\"%s\\\":\\\"%s\\\"},\\\"%s\\\":{\\\"%s\\\":\\\"%s\\\", \\\"%s\\\":\\\"%s\\\"}}}," +
                                    "{\\\"op\\\":\\\"replace\\\"," +
                                    "\\\"path\\\":\\\"/spec/replicas\\\"," +
                                    "\\\"value\\\": %d}" +
                                    "]\" -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",
                            configs.get(0).getType(),configs.get(0).getValues().get(0).getKey(),configs.get(0).getValues().get(0).getValue(),configs.get(0).getValues().get(1).getKey(),configs.get(0).getValues().get(1).getValue(),
                            configs.get(1).getType(),configs.get(1).getValues().get(0).getKey(),configs.get(1).getValues().get(0).getValue(),configs.get(1).getValues().get(1).getKey(),configs.get(1).getValues().get(1).getValue(),
                            request.getNumOfReplicas(),
                            apiUrl,cluster.getToken(),filePath)
                    };
                }
                //Delta limits and requests only
                else{
                    System.out.println("[Delta All] Delta limits and requests only");
                    cmds = new String[]{
                            "/bin/sh","-c",String.format("curl -X PATCH -d \"[" +
                                    "{\\\"op\\\":\\\"replace\\\"," +
                                    "\\\"path\\\":\\\"/spec/template/spec/containers/0/resources\\\"," +
                                    "\\\"value\\\":{\\\"%s\\\":{\\\"%s\\\":\\\"%s\\\", \\\"%s\\\":\\\"%s\\\"},\\\"%s\\\":{\\\"%s\\\":\\\"%s\\\", \\\"%s\\\":\\\"%s\\\"}}}" +
                                    "]\" -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",
                            configs.get(0).getType(),configs.get(0).getValues().get(0).getKey(),configs.get(0).getValues().get(0).getValue(),configs.get(0).getValues().get(1).getKey(),configs.get(0).getValues().get(1).getValue(),
                            configs.get(1).getType(),configs.get(1).getValues().get(0).getKey(),configs.get(1).getValues().get(0).getValue(),configs.get(1).getValues().get(1).getKey(),configs.get(1).getValues().get(1).getValue(),
                            apiUrl,cluster.getToken(),filePath)
                    };
                }
            }
        }

        System.out.println(String.format("The constructed command for deltaing all is %s", cmds[2]));
        if(cmds.length > 0){
            ProcessBuilder pb = new ProcessBuilder(cmds);
            pb.redirectErrorStream(true);
            Process p;
            try {
                p = pb.start();
                p.waitFor();
                Thread.sleep(2000);
                String json = readWholeFile(filePath);
                //Parse the response to the SetServicesReplicasResponseFromAPI Bean
//            System.out.println(json);
                result = JSON.parseObject(json,SingleDeploymentInfo.class);
            } catch (Exception e) {
                isSuccess = false;
                e.printStackTrace();
            }
        }else{
            System.out.println("The cmds is empty. No command to execute!");
        }

        return isSuccess;
    }

    //Delete the specified pod
    private boolean deletePod(String namespace, String podName, Cluster cluster){
        boolean isSuccess = true;
        V1Pod result;
        String filePath = "/app/delete_pod_result_" + cluster.getName() + System.currentTimeMillis()+ ".json";
        String apiUrl = String.format("%s/api/v1/namespaces/%s/pods/%s",cluster.getApiServer(),namespace, podName);
        System.out.println(String.format("The constructed api url for deleting node is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X DELETE %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,cluster.getToken(),filePath)
        };
        ProcessBuilder pb = new ProcessBuilder(cmds);
        pb.redirectErrorStream(true);
        Process p;
        try {
            p = pb.start();
            p.waitFor();

            String json = readWholeFile(filePath);
            //Parse the response to the SetServicesReplicasResponseFromAPI Bean
//            System.out.println(json);
            result = JSON.parseObject(json,V1Pod.class);
        } catch (Exception e) {
            isSuccess = false;
            e.printStackTrace();
        }
        return isSuccess;
    }

    //Get the logs of a named pod
    private String getPodLog(String podName,String containerName, Cluster cluster){
        String log = "";

        String filePath = "/app/get_pod_log_" + cluster.getName() + System.currentTimeMillis()+ ".json";
        String apiUrl = String.format("%s/api/v1/namespaces/%s/pods/%s/log?container=%s",cluster.getApiServer(),NAMESPACE,podName,containerName);
        System.out.println(String.format("The constructed api url for getting the pod log is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,cluster.getToken(),filePath)
        };
        ProcessBuilder pb = new ProcessBuilder(cmds);
        pb.redirectErrorStream(true);
        Process p;
        try {
            p = pb.start();
            p.waitFor();
            log = readWholeFile(filePath);
        } catch (IOException e) {
            e.printStackTrace();
        }catch(InterruptedException e){
            e.printStackTrace();
        }
        return log;
    }

    //To determine if the service need to be deleted: filter the redis and mongo
    private boolean isDeleted(String deploymentName,List<String> serviceNames){
        return (deploymentName.contains("service")) && !existInTheList(deploymentName,serviceNames);
    }

    //Judge if the node name is in the reserved node name list
    private boolean isExistInNodeList(String targetNodeName, List<String> nodeNames){
        boolean isExist = false;
        for(String nodeName : nodeNames){
            if(targetNodeName.equals(nodeName)){
                isExist = true;
                break;
            }
        }
        return isExist;
    }

    //Set service to the target replicas number
    private boolean setServiceReplica(String serviceName, int targetNum, Cluster cluster){
        boolean status = false;
        SetServicesReplicasResponseFromAPI result;
        String filePath = "/app/set_service_replica_" + cluster.getName() + System.currentTimeMillis()+ ".json";
        String apiUrl = String.format("%s/apis/extensions/v1beta1/namespaces/%s/deployments/%s/scale",cluster.getApiServer() ,NAMESPACE,serviceName);
        System.out.println(String.format("The constructed api url is %s", apiUrl));
        String data ="'[{ \"op\": \"replace\", \"path\": \"/spec/replicas\", \"value\":" +  targetNum + " }]'";

        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X PATCH -d%s -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",data,apiUrl,cluster.getToken(),filePath)
        };
//        System.out.println(String.format("The constructed command for setting service replicas is %s", cmds[2]));
        ProcessBuilder pb = new ProcessBuilder(cmds);
        pb.redirectErrorStream(true);
        Process p;
        try {
            p = pb.start();
            p.waitFor();

            String json = readWholeFile(filePath);
            //Parse the response to the SetServicesReplicasResponseFromAPI Bean
//            System.out.println(json);
            result = JSON.parseObject(json,SetServicesReplicasResponseFromAPI.class);
            status = true;
            System.out.println(String.format("The pod corresponding to service %s has been deleted successfully!",serviceName));
        } catch (Exception e) {
            status = false;
            System.out.println(String.format("Fail to delete the pod corresponding to service %s",serviceName));
            e.printStackTrace();
        }
        return status;
    }

    //Judge if the deployment name is in the reserved service name list
    private boolean existInTheList(String deploymentName, List<String> serviceNames){
        boolean isExist = false;
        for(String serviceName : serviceNames){
            if(deploymentName.equals(serviceName)){
                isExist = true;
                break;
            }
        }
        return isExist;
    }

    //Delete the node
    private boolean deleteNode(String nodeName, Cluster cluster){
        boolean isSuccess = true;
        DeleteNodeResult result;
        String filePath = "/app/delete_node_result_" + cluster.getName() + System.currentTimeMillis()+ ".json";
        String apiUrl = String.format("%s/api/v1/nodes/%s",cluster.getApiServer(),nodeName );
        System.out.println(String.format("The constructed api url for deleting node is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X DELETE %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,cluster.getToken(),filePath)
        };
        ProcessBuilder pb = new ProcessBuilder(cmds);
        pb.redirectErrorStream(true);
        Process p;
        try {
            p = pb.start();
            p.waitFor();

            String json = readWholeFile(filePath);
            //Parse the response to the SetServicesReplicasResponseFromAPI Bean
//            System.out.println(json);
            result = JSON.parseObject(json,DeleteNodeResult.class);
            if(result.getStatus().equals("Success") || result.getStatus().equals("success")){
                System.out.println(String.format("The node %s has been deleted successfully!",nodeName));
            }
            else{
                isSuccess = false;
                System.out.println(String.format("Fail to delete the node %s. The corresponding status is %s",nodeName,result.getStatus()));
            }
        } catch (IOException e) {
            e.printStackTrace();
        }catch(InterruptedException e){
            e.printStackTrace();
        }
        return isSuccess;
    }

    //Delete the service
    private void deleteService(String serviceName, Cluster cluster){
        DeleteServiceResult result;
        String filePath = "/app/delete_service_result_" + cluster.getName() + System.currentTimeMillis()+ ".json";
        String apiUrl = String.format("%s/api/v1/namespaces/%s/services/%s",cluster.getApiServer(), NAMESPACE,serviceName );
        System.out.println(String.format("The constructed api url for deleting service is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X DELETE %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,cluster.getToken(),filePath)
        };
        ProcessBuilder pb = new ProcessBuilder(cmds);
        pb.redirectErrorStream(true);
        Process p;
        try {
            p = pb.start();
            p.waitFor();

            String json = readWholeFile(filePath);
            //Parse the response to the SetServicesReplicasResponseFromAPI Bean
//            System.out.println(json);
            result = JSON.parseObject(json,DeleteServiceResult.class);
            if(result.getStatus().equals("Success") || result.getStatus().equals("success")){
                System.out.println(String.format("The service %s has been deleted successfully!",serviceName));
            }
            else{
                System.out.println(String.format("Fail to delete the service %s. The corresponding status is %s",serviceName,result.getStatus()));
            }
        } catch (IOException e) {
            e.printStackTrace();
        }catch(InterruptedException e){
            e.printStackTrace();
        }
    }

    //Check if all the services are available again after deleting the node
    private boolean isAllReady(String namespace, Cluster cluster){
        boolean isAllReady = true;
        QueryDeploymentsListResponse deploymentsList = getDeploymentList(namespace, cluster);
        for(SingleDeploymentInfo singleDeploymentInfo : deploymentsList.getItems()){
            if(singleDeploymentInfo.getStatus().getReplicas() != singleDeploymentInfo.getStatus().getReadyReplicas()){
                isAllReady = false;
                break;
            }
        }
        return isAllReady;
    }

    //Check if all the services are able to serve
    private boolean isAllServiceReadyToServe(String namespace, Cluster cluster){
        System.out.println("Check if all the services are able to serve");
        boolean result;
        QueryDeploymentsListResponse deploymentsList = getDeploymentList(namespace, cluster);
        List<String> serviceNames = new ArrayList<>();
        for(SingleDeploymentInfo singleDeploymentInfo : deploymentsList.getItems()){
            String serviceName = singleDeploymentInfo.getMetadata().getName();
            if(serviceName.contains("service")){
                serviceNames.add(serviceName);
            }
        }
        System.out.println(String.format("The service list to be checked is %s", serviceNames.toString()));
        result = isAllAbleToServe(serviceNames, cluster);
        return result;
    }

    //Check if all the required deployment replicas are ready
    private boolean isAllReady(SetServiceReplicasRequest setServiceReplicasRequest, Cluster cluster){
        boolean isAllReady = true;

        QueryDeploymentsListResponse deploymentsList = getDeploymentList(NAMESPACE, cluster);

        for(ServiceReplicasSetting setting : setServiceReplicasRequest.getServiceReplicasSettings()){
            if(!isSingleReady(deploymentsList.getItems(),setting)){
                isAllReady = false;
                break;
            }
        }
        return isAllReady;
    }

    //Get the deployment list
    private QueryDeploymentsListResponse getDeploymentList(String namespace, Cluster cluster){
        //Get the current deployments information and echo to the file
        String filePath = "/app/get_deployment_list_result_" + cluster.getName()  + System.currentTimeMillis()+ ".json";
        QueryDeploymentsListResponse deploymentsList = new QueryDeploymentsListResponse();
        String apiUrl = String.format("%s/apis/apps/v1beta1/namespaces/%s/deployments",cluster.getApiServer() ,namespace);
        System.out.println(String.format("The constructed api url for getting the deploymentlist is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,cluster.getToken(),filePath)
        };
        ProcessBuilder pb = new ProcessBuilder(cmds);
        pb.redirectErrorStream(true);
        Process p;
        try {
            p = pb.start();
            p.waitFor();

            //Wait 2 seconds to ensure the existence of the file
            Thread.sleep(5000);

            String json = readWholeFile(filePath);
            //Parse the response to the SetServicesReplicasResponseFromAPI Bean
//            System.out.println(json);
            deploymentsList = JSON.parseObject(json,QueryDeploymentsListResponse.class);
        } catch (Exception e) {
            e.printStackTrace();
        }
        return deploymentsList;
    }

    //Get the node list
    private V1NodeList getNodeList(Cluster cluster){
        //Get the current deployments information and echo to the file
        String filePath = "/app/get_node_list_result_" + cluster.getName() + System.currentTimeMillis() + ".json";
        V1NodeList nodeList = new V1NodeList();
        String apiUrl = String.format("%s/api/v1/nodes",cluster.getApiServer());
        System.out.println(String.format("The constructed api url for getting the node list is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,cluster.getToken(),filePath)
        };

        ProcessBuilder pb = new ProcessBuilder(cmds);
        pb.redirectErrorStream(true);
        Process p;
        try {
            p = pb.start();
            p.waitFor();

            String json = readWholeFile(filePath);
            //Parse the response to the SetServicesReplicasResponseFromAPI Bean
//            System.out.println(json);
            nodeList = JSON.parseObject(json,V1NodeList.class);
        } catch (IOException e) {
            e.printStackTrace();
        }catch(InterruptedException e){
            e.printStackTrace();
        }
        return nodeList;
    }

    //Get the pods list
    private V1PodList getPodList(String namespace, Cluster cluster){
        if(namespace.equals(""))
            namespace = NAMESPACE;
        //Get the current pods information and echo to the file
        String filePath = "/app/get_pod_list_result_" + cluster.getName() + System.currentTimeMillis()+ ".json";
        V1PodList podList = new V1PodList();
        String apiUrl = String.format("%s/api/v1/namespaces/%s/pods",cluster.getApiServer(),namespace);
        System.out.println(String.format("The constructed api url for getting the pod list is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,cluster.getToken(),filePath)
        };
        ProcessBuilder pb = new ProcessBuilder(cmds);
        pb.redirectErrorStream(true);
        Process p;
        try {
            p = pb.start();
            p.waitFor();

            String json = readWholeFile(filePath);
            //Parse the response to the SetServicesReplicasResponseFromAPI Bean
//            System.out.println(json);
            podList = JSON.parseObject(json,V1PodList.class);
        } catch (IOException e) {
            e.printStackTrace();
        }catch(InterruptedException e){
            e.printStackTrace();
        }
        return podList;
    }

    //Get the pod info
    private V1Pod getPodInfo(String name, Cluster cluster){
        //Get the current pods information and echo to the file
        String filePath = "/app/get_pod_info_result_"+ cluster.getName()  + System.currentTimeMillis()+".json";
        V1Pod pod = null;
        String apiUrl = String.format("%s/api/v1/namespaces/%s/pods/%s",cluster.getApiServer(),NAMESPACE,name);
        System.out.println(String.format("The constructed api url for getting the pod info is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,cluster.getToken(),filePath)
        };
        ProcessBuilder pb = new ProcessBuilder(cmds);
        pb.redirectErrorStream(true);
        Process p;
        try {
            p = pb.start();
            p.waitFor();

            String json = readWholeFile(filePath);
            //Parse the response to the SetServicesReplicasResponseFromAPI Bean
//            System.out.println(json);
            pod = JSON.parseObject(json,V1Pod.class);
        } catch (Exception e) {
            e.printStackTrace();
        }
        return pod;
    }

    //Check if the single required deployment replicas are ready
    private boolean isSingleReady(List<SingleDeploymentInfo> deploymentInfoList, ServiceReplicasSetting setting){
        boolean isReady = false;
        for(SingleDeploymentInfo singleDeploymentInfo : deploymentInfoList){
            if(singleDeploymentInfo.getMetadata().getName().equals(setting.getServiceName())){
                System.out.println(String.format("The desired replicas of service %s is %d, the ready number of replicas is %d", setting.getServiceName(),setting.getNumOfReplicas(),singleDeploymentInfo.getStatus().getReadyReplicas()));
                if(singleDeploymentInfo.getStatus().getReadyReplicas() == setting.getNumOfReplicas()){
                    isReady = true;
                    System.out.println(String.format("The service %s has already set the required number of replicas", setting.getServiceName()));
                    break;
                }
            }
        }
        return isReady;
    }

    //Read the whole file(Delete after read completion)
    private String readWholeFile(String path){
        String encoding = "UTF-8";
        File file = new File(path);
        Long filelength = file.length();
        byte[] filecontent = new byte[filelength.intValue()];
        try {
            FileInputStream in = new FileInputStream(file);
            in.read(filecontent);
            in.close();
        } catch (FileNotFoundException e) {
            e.printStackTrace();
        } catch (IOException e) {
            e.printStackTrace();
        }
        try {
            return new String(filecontent, encoding);
        } catch (UnsupportedEncodingException e) {
            System.err.println("The OS does not support " + encoding);
            e.printStackTrace();
            return null;
        }finally {
            deleteFile(path);
        }
    }

    //Delete the temporary file
    private void deleteFile(String filePath){
        try {
            File file = new File(filePath);
            if (file.delete()) {
                System.out.println(file.getName() + " is deleted");
            } else {
                System.out.println("Delete failed.");
            }
        } catch (Exception e) {
            System.out.println("Exception occured when delete file");
            e.printStackTrace();
        }
    }
}
