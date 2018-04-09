package apiserver.service;

import apiserver.bean.*;
import apiserver.request.*;
import apiserver.response.*;
import apiserver.util.FileOperation;
import apiserver.util.MyConfig;
import apiserver.util.RemoteExecuteCommand;
import com.alibaba.fastjson.JSON;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import java.io.*;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import ch.ethz.ssh2.SFTPv3Client;

@Service
public class ApiServiceImpl implements ApiService {

    private final String NAMESPACE = "default";

    private String masterIp = "10.141.212.21";

    private String username = "root";

    private String password = "root";

    @Autowired
    private MyConfig myConfig;

    @Override
    public SetUnsetServiceRequestSuspendResponse setServiceRequestSuspend(SetUnsetServiceRequestSuspendRequest setUnsetServiceRequestSuspendRequest){

        String svcName = setUnsetServiceRequestSuspendRequest.getSvc();
        String executeResult = doSetServiceRequestSuspend(svcName);
        System.out.println(executeResult);
        boolean status = (executeResult != null);
        SetUnsetServiceRequestSuspendResponse response = new SetUnsetServiceRequestSuspendResponse(status,executeResult);
        return response;
    }

    @Override
    public SetUnsetServiceRequestSuspendResponse setServiceRequestSuspendWithSource(SetUnsetServiceRequestSuspendRequest setUnsetServiceRequestSuspendRequest){

        String svcName = setUnsetServiceRequestSuspendRequest.getSvc();
        String sourceSvcName = setUnsetServiceRequestSuspendRequest.getSourceSvcName();
        String executeResult = doSetServiceRequestSuspendWithSourceFile(svcName,sourceSvcName);
        System.out.println(executeResult);
        boolean status = (executeResult != null);
        SetUnsetServiceRequestSuspendResponse response = new SetUnsetServiceRequestSuspendResponse(status,executeResult);
        return response;
    }

    private String doSetServiceRequestSuspend(String svcName){
        String svcLongDelayFilePath = "rule-long-" + svcName + ".yml";

//        FileOperation.clearAndWriteFile(svcLongDelayFilePath,svcName);
        RemoteExecuteCommand rec = new RemoteExecuteCommand(masterIp, username,password);
        rec.modifyFile(svcLongDelayFilePath,svcName);
//        rec.uploadFile(svcLongDelayFilePath);

        String serLongDelayRequest = "kubectl apply -f " + svcLongDelayFilePath;
        //执行脚本
        String executeResult = rec.execute("export KUBECONFIG=/etc/kubernetes/admin.conf;" + serLongDelayRequest);
        return executeResult;
    }

    private String doSetServiceRequestSuspendWithSourceFile(String svcName, String sourceSvcName){
        String svcLongDelayFilePath = "rule-long-" + svcName + ".yml";

//        FileOperation.clearAndWriteFile(svcLongDelayFilePath,svcName);
        RemoteExecuteCommand rec = new RemoteExecuteCommand(masterIp, username,password);
        rec.modifyFileWithSourceSvcName(svcLongDelayFilePath,svcName,sourceSvcName);
//        rec.uploadFile(svcLongDelayFilePath);

        String serLongDelayRequest = "kubectl apply -f " + svcLongDelayFilePath;
        //执行脚本
        String executeResult = rec.execute("export KUBECONFIG=/etc/kubernetes/admin.conf;" + serLongDelayRequest);
        return executeResult;
    }


    @Override
    public SetUnsetServiceRequestSuspendResponse unsetServiceRequestSuspend(SetUnsetServiceRequestSuspendRequest setUnsetServiceRequestSuspendRequest){
        String svcName = setUnsetServiceRequestSuspendRequest.getSvc();
        String executeResult = doUnsetServiceRequestSuspend(svcName);
        System.out.println(executeResult);
        boolean status = (executeResult != null);
        SetUnsetServiceRequestSuspendResponse response = new SetUnsetServiceRequestSuspendResponse(status,executeResult);
        return response;
    }

    private String doUnsetServiceRequestSuspend(String svcName){
        String svcLongDelayFilePath = "rule-long-" + svcName + ".yml";
        String serLongDelayRequest = "kubectl delete -f " + svcLongDelayFilePath;
        RemoteExecuteCommand rec = new RemoteExecuteCommand(masterIp, username,password);
        //执行脚本
        return rec.execute("export KUBECONFIG=/etc/kubernetes/admin.conf;" + serLongDelayRequest);
    }

    private String applyYml(String fileName){
        String applyRequest = "kubectl apply -f " + fileName;
        RemoteExecuteCommand rec = new RemoteExecuteCommand(masterIp, username,password);
        //执行脚本
        return rec.execute("export KUBECONFIG=/etc/kubernetes/admin.conf;" + applyRequest);
    }

    private String deleteYml(String fileName){
        String applyRequest = "kubectl delete -f " + fileName;
        RemoteExecuteCommand rec = new RemoteExecuteCommand(masterIp, username,password);
        //执行脚本
        return rec.execute("export KUBECONFIG=/etc/kubernetes/admin.conf;" + applyRequest);
    }

    @Override
    public SetAsyncRequestSequenceResponse setAsyncRequestsSequence(SetAsyncRequestSequenceRequest setAsyncRequestSequenceRequest){
        ArrayList<String> svcList = setAsyncRequestSequenceRequest.getSvcList();
        for(int i = 0;i < svcList.size(); i++){
            String svcName = svcList.get(i);
            System.out.println("[=====]释放 " + svcName + ": " + doUnsetServiceRequestSuspend(svcName));
            //waitForComplete是阻塞式的 会一直等待直到请求返回
            if(waitForComplete(svcName) == true) {
                System.out.println("[===== Complete =====] " + svcName);
            }
        }
        SetAsyncRequestSequenceResponse request = new SetAsyncRequestSequenceResponse(true," setAsyncRequestsSequence Complete");
        return request;
    }

    private boolean waitForComplete(String svcName){
        //根据svc的名称，获取svc下的所有pod
        GetPodsListResponse podsListResponse = getPodsList("default");
        ArrayList<PodInfo> podInfoList = new ArrayList<>(podsListResponse.getPods());
        ArrayList<PodInfo> targetPodInfoList = new ArrayList<>();
        for(PodInfo podInfo : podInfoList){
            System.out.println("[=====] We are now checking useful POD-NAME:" + podInfo.getName());
            if(podInfo.getName().contains(svcName)){
                targetPodInfoList.add(podInfo);
            }else{
                //do nothing
            }
        }
        boolean isRequestComplete = false;
//        try{
//            Thread.sleep(90000);
//            isRequestComplete = true;
//        }catch (Exception e){
//            e.printStackTrace();
//        }
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
                String podLog = getPodLog(podInfo.getName(),"istio-proxy");
                String[] logsFormatted = podLog.split("\n");
                ArrayList<String> arrayList = new ArrayList<>(Arrays.asList(logsFormatted));
                ArrayList<String> lastSeveralLogs = new ArrayList<>(arrayList.subList(arrayList.size() - 10,arrayList.size()));

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
        SetServiceReplicasResponse response = new SetServiceReplicasResponse();
        //Set the desired number of service replicas
        for(ServiceReplicasSetting setting : setServiceReplicasRequest.getServiceReplicasSettings()){
            String apiUrl = String.format("%s/apis/extensions/v1beta1/namespaces/%s/deployments/%s/scale",myConfig.getApiServer() ,NAMESPACE,setting.getServiceName());
            System.out.println(String.format("The constructed api url is %s", apiUrl));
            String data ="'[{ \"op\": \"replace\", \"path\": \"/spec/replicas\", \"value\":" +  setting.getNumOfReplicas() + " }]'";

            String[] cmds ={
                    "/bin/sh","-c",String.format("curl -X PATCH -d%s -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure",data,apiUrl,myConfig.getToken())
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
        //Check if all the required replicas are ready
        while(!isAllReady(setServiceReplicasRequest)){
            try{
                //Check every 10 seconds
                Thread.sleep(10000);
            }catch(Exception e){
                e.printStackTrace();
            }
        }
        response.setStatus(true);
        response.setMessage("All the required service replicas have been already set!");
        return response;
    }

    //Get all of the services name
    @Override
    public GetServicesListResponse getServicesList() {
        GetServicesListResponse response = new GetServicesListResponse();
        //Get the current deployments information
        QueryDeploymentsListResponse deploymentsList = getDeploymentList(NAMESPACE);
        //Iterate the list and return the result
        List<ServiceWithReplicas> services = new ArrayList<ServiceWithReplicas>();
        for(SingleDeploymentInfo singleDeploymentInfo : deploymentsList.getItems()){
            ServiceWithReplicas serviceWithReplicas = new ServiceWithReplicas();
            serviceWithReplicas.setServiceName(singleDeploymentInfo.getMetadata().getName());
            serviceWithReplicas.setNumOfReplicas(singleDeploymentInfo.getStatus().getReadyReplicas());
            services.add(serviceWithReplicas);
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
        GetServiceReplicasResponse response = new GetServiceReplicasResponse();
        //Get the current deployments information
        QueryDeploymentsListResponse deploymentsList = getDeploymentList(NAMESPACE);
        //Iterate the list and return the result
        List<ServiceWithReplicas> services = new ArrayList<ServiceWithReplicas>();
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
        ReserveServiceByListResponse response = new ReserveServiceByListResponse();
        response.setStatus(true);
        response.setMessage("Succeed to delete all of the services not contained in the list");
        //Get the current deployments information
        QueryDeploymentsListResponse deploymentsList = getDeploymentList(NAMESPACE);

        for(SingleDeploymentInfo singleDeploymentInfo : deploymentsList.getItems()){
            //Delete the services not contained in the list
            String deploymentName = singleDeploymentInfo.getMetadata().getName();
            if(isDeleted(deploymentName,reserveServiceRequest.getServices())){
                System.out.println(String.format("The service %s isn't contained in the reserved list. To be deleted",deploymentName ));
                //Delete the service first
                deleteService(deploymentName);
                //Delete the corresponding pod by set the number of replica to 0
                boolean result = setServiceReplica(deploymentName, 0);
                if(!result){
                    response.setStatus(false);
                    response.setMessage(String.format("Fail to delete the service %s", deploymentName));
                    break;
                }
            }else{
                System.out.println(String.format("The service %s is contained in the reserved list. Reserve",deploymentName ));
            }
        }
        return response;
    }

    //Set the system to run on single node
    @Override
    public SetRunOnSingleNodeResponse setRunOnSingleNode() {
        SetRunOnSingleNodeResponse response = new SetRunOnSingleNodeResponse();
        //Set the default information
        response.setStatus(false);
        response.setMessage("There is no message now!");

        V1NodeList nodeList = getNodeList();
        List<V1Node> workingNodeList = new ArrayList<V1Node>();

        //Construct the working node list
        for(V1Node node : nodeList.getItems()){
            System.out.println(String.format("The node name is %s and the role is %s",node.getMetadata().getName(),node.getSpec().getTaints() == null?"Minion":"Master"));
            if(node.getSpec().getTaints() == null)
                workingNodeList.add(node);
        }

        //Delete the other working nodes and reserve only one and return until all the services are available again
        if(workingNodeList.size() <= 1){
            System.out.println("There is at most one working node. Nothing to do.");
        }else{
            //Delete node
            for(int i = 1; i < workingNodeList.size(); i++){
                V1Node node = workingNodeList.get(i);
                System.out.println(String.format("The node %s is to be deleted",node.getMetadata().getName()));
                deleteNode(node.getMetadata().getName());
            }
            //Sleep to let the kubernetes has time to refresh the status
            //TODO: time value to be determined. It seems that 10s is too short。 30s is enough
            try{
                Thread.sleep(30000);
            }catch(Exception e){
                e.printStackTrace();
            }
            //Check whether all of the services are available again in certain internals
            while(!isAllReady(NAMESPACE)){
                try{
                    //Check every 10 seconds
                    Thread.sleep(10000);
                }catch(Exception e){
                    e.printStackTrace();
                }
            }
            response.setStatus(true);
            response.setMessage("The system are now run on single node");
        }
        return response;
    }

    //Get the node list
    @Override
    public GetNodesListResponse getNodesList() {
        GetNodesListResponse response = new GetNodesListResponse();
        V1NodeList nodeList = getNodeList();
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
        DeltaNodeByListResponse response = new DeltaNodeByListResponse();
        List<String> nodeNames = deltaNodeRequest.getNodeNames();
        boolean isSuccess =true;
        for(String nodeName : nodeNames){
            if(!deleteNode(nodeName))
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
        while(!isAllReady(NAMESPACE)){
            try{
                //Check every 10 seconds
                Thread.sleep(10000);
            }catch(Exception e){
                e.printStackTrace();
            }
        }
        return response;
    }

    //Reserve the nodes contained in the list
    @Override
    public DeltaNodeByListResponse reserveNodeByList(DeltaNodeRequest deltaNodeRequest) {
        DeltaNodeByListResponse response = new DeltaNodeByListResponse();
        List<String> nodeNames = deltaNodeRequest.getNodeNames();
        V1NodeList nodeList = getNodeList();
        boolean isSuccess =true;
        for(V1Node node : nodeList.getItems()){
            if(node.getSpec().getTaints() != null){
                System.out.println("The master can't be deleted!");
                continue;
            }
            String nodeName = node.getMetadata().getName();
            if(!isExistInNodeList(nodeName,nodeNames)){
                if(!deleteNode(nodeName))
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
        while(!isAllReady(NAMESPACE)){
            try{
                //Check every 10 seconds
                Thread.sleep(10000);
            }catch(Exception e){
                e.printStackTrace();
            }
        }
        return response;
    }

    //Get the pods info list
    @Override
    public GetPodsListResponse getPodsList() {
        GetPodsListResponse response = new GetPodsListResponse();
        V1PodList podList = getPodList("");
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

    public GetPodsListResponse getPodsList(String namespace) {
        GetPodsListResponse response = new GetPodsListResponse();
        V1PodList podList = getPodList(namespace);
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
    public GetPodsLogResponse getPodsLog() {
        GetPodsLogResponse response = new GetPodsLogResponse();
        V1PodList podList = getPodList("");
        System.out.println(String.format("There are now %d pods in the cluster now", podList.getItems().size()));
        if(podList.getItems().size() < 1){
            response.setStatus(true);
            response.setMessage("No resource found!");
            response.setPodLogs(null);
        }
        //Construct the pods log info
        List<PodLog> podLogs = new ArrayList<PodLog>();
        List<V1Container> containers;
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
            String logs = getPodLog(podName,containerName);
            podLog.setLogs(logs);
            podLogs.add(podLog);
        }
        response.setStatus(true);
        response.setMessage("Successfully to get the pods log info!");
        response.setPodLogs(podLogs);
        return response;
    }

    //Get the log of the single pod
    @Override
    public GetSinglePodLogResponse getSinglePodLog(GetSinglePodLogRequest getSinglePodLogRequest) {
        GetSinglePodLogResponse response = new GetSinglePodLogResponse();
        response.setStatus(false);
        response.setMessage("Fail to get the corresponding pod's log!");
        response.setPodLog(null);

        //Get the pod name and the container name
        PodLog podLog = new PodLog();
        String podName = getSinglePodLogRequest.getPodName();
        String containerName = "";
        V1Pod pod = getPodInfo(podName);
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
            String log = getPodLog(podName,containerName);
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
    public RestartServiceResponse restartService() {
        boolean isExist = false;
        RestartServiceResponse response = new RestartServiceResponse();
        V1PodList podList = getPodList("istio-system");
        //Delete the zipkin pod
        for(V1Pod pod : podList.getItems()) {
            String podName = pod.getMetadata().getName();
            if(podName.contains("zipkin")){
                isExist = true;
                //Delete the pod corresponding to the service
                boolean deleteResult = deletePod("istio-system",podName);
                if(!deleteResult){
                    response.setStatus(false);
                    response.setMessage("Fail to restart zipkin!");
                    return response;
                }
                break;
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
            while(!isAllReady("istio-system")){
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
    public GetServicesAndConfigResponse getServicesAndConfig() {
        GetServicesAndConfigResponse response = new GetServicesAndConfigResponse();
        //Get the current deployments information
        QueryDeploymentsListResponse deploymentsList = getDeploymentList(NAMESPACE);
        //Iterate the list and return the result
        List<ServiceWithConfig> services = new ArrayList<ServiceWithConfig>();
        for(SingleDeploymentInfo singleDeploymentInfo : deploymentsList.getItems()){
            V1ResourceRequirements resourceRequirements = singleDeploymentInfo.getSpec().getTemplate().getSpec().getContainers().get(0).getResources();
            ServiceWithConfig serviceWithConfig = new ServiceWithConfig();
            serviceWithConfig.setServiceName(singleDeploymentInfo.getMetadata().getName());
            serviceWithConfig.setLimits(resourceRequirements.getLimits());
            serviceWithConfig.setRequests(resourceRequirements.getRequests());
            services.add(serviceWithConfig);
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
        DeltaCMResourceResponse response = new DeltaCMResourceResponse();
        //Get the current deployments information
        QueryDeploymentsListResponse deploymentsList = getDeploymentList(NAMESPACE);

        //Check if the resource setting exists
        for(SingleDeltaCMResourceRequest request : deltaCMResourceRequest.getDeltaRequests()){
            for(SingleDeploymentInfo singleDeploymentInfo : deploymentsList.getItems()){
                if(singleDeploymentInfo.getMetadata().getName().equals(request.getServiceName())){
                    V1ResourceRequirements resourceRequirements = singleDeploymentInfo.getSpec().getTemplate().getSpec().getContainers().get(0).getResources();
                    if(resourceRequirements.getLimits() == null || resourceRequirements.getRequests() == null){
                        response.setMessage("There is no corresponding config in the cluster!");
                        response.setStatus(false);
                        //TODO: Add the config through add option
                    }
                    else{
                        boolean result = deltaCMResource(NAMESPACE,request);
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
        while(!isAllReady(NAMESPACE)){
            try{
                //Check every 2 seconds
                Thread.sleep(2000);
            }catch(Exception e){
                e.printStackTrace();
            }
        }

        return response;
    }

    //Get the service with endpoints
    @Override
    public ServiceWithEndpointsResponse getServiceWithEndpoints() {
        ServiceWithEndpointsResponse response = new ServiceWithEndpointsResponse();
        //Get the current endpoints list
        V1EndpointsList endpointsList = getEndpointsList(NAMESPACE);
        //Iterate the list and return the result
        List<ServiceWithEndpoints> services = new ArrayList<ServiceWithEndpoints>();
        for(V1Endpoints endpoints : endpointsList.getItems()){
            ServiceWithEndpoints serviceWithEndpoints = new ServiceWithEndpoints();
            //Set service name
            serviceWithEndpoints.setServiceName(endpoints.getMetadata().getName());
            //Set service endpoints
            List<String> endpointsListOfService = new ArrayList<>();
            if(endpoints.getSubsets().size() > 0){
                V1EndpointSubset subset = endpoints.getSubsets().get(0);
                for(V1EndpointAddress v1EndpointAddress : subset.getAddresses()){
                    String ip = v1EndpointAddress.getIp();
                    for(V1EndpointPort v1EndpointPort : subset.getPorts()){
                        endpointsListOfService.add(ip + ":" + v1EndpointPort.getPort());
                    }
                }
                serviceWithEndpoints.setEndPoints(endpointsListOfService);
            }
            services.add(serviceWithEndpoints);
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
        ServiceWithEndpointsResponse response = new ServiceWithEndpointsResponse();
        List<ServiceWithEndpoints> services = new ArrayList<ServiceWithEndpoints>();
        for(String serviceName : reserveServiceRequest.getServices()){
            V1Endpoints endpoints = getSingleServiceEndpoints(NAMESPACE,serviceName);
            ServiceWithEndpoints serviceWithEndpoints = new ServiceWithEndpoints();
            //Set service name
            serviceWithEndpoints.setServiceName(endpoints.getMetadata().getName());
            //Set service endpoints
            List<String> endpointsListOfService = new ArrayList<>();
            if(endpoints.getSubsets().size() > 0){
                V1EndpointSubset subset = endpoints.getSubsets().get(0);
                for(V1EndpointAddress v1EndpointAddress : subset.getAddresses()){
                    String ip = v1EndpointAddress.getIp();
                    for(V1EndpointPort v1EndpointPort : subset.getPorts()){
                        endpointsListOfService.add(ip + ":" + v1EndpointPort.getPort());
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

    //Get the endpoints list of specific service
    private V1Endpoints getSingleServiceEndpoints(String namespace, String serviceName){
        V1Endpoints endpoints = new V1Endpoints();
        String filePath = "/app/get_endpoints_list_of_single_service.json";
        String apiUrl = String.format("%s/api/v1/namespaces/%s/endpoints/%s",myConfig.getApiServer(),namespace,serviceName);
        System.out.println(String.format("The constructed api url for getting the endpoints of specific service is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,myConfig.getToken(),filePath)
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
    private V1EndpointsList getEndpointsList(String namespace){
        V1EndpointsList endpointsList = new V1EndpointsList();
        String filePath = "/app/get_endpoints_list.json";
        String apiUrl = String.format("%s/api/v1/namespaces/%s/endpoints",myConfig.getApiServer(),namespace);
        System.out.println(String.format("The constructed api url for getting the endpoints list is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,myConfig.getToken(),filePath)
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

    //Delta the CPU and memory of service
    private boolean deltaCMResource(String namespace,SingleDeltaCMResourceRequest request){
        boolean isSuccess = true;
        SingleDeploymentInfo result;
        String filePath = "/app/delta_cmconfig_result.json";
        String apiUrl = String.format("%s/apis/apps/v1beta1/namespaces/%s/deployments/%s",myConfig.getApiServer(),namespace, request.getServiceName());
        System.out.println(String.format("The constructed api url for deltaing config is %s", apiUrl));
        String data = String.format("'[{ \"op\": \"replace\", \"path\": \"/spec/template/spec/containers/0/resources/%s/%s\", \"value\": \"%s\"}]'",
                request.getType(),request.getKey(),request.getValue());
        System.out.println(String.format("The constructed data for deltaing config is %s", data));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X PATCH -d%s -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",
                data,apiUrl,myConfig.getToken(),filePath)
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

    //Delete the specified pod
    private boolean deletePod(String namespace, String podName){
        boolean isSuccess = true;
        V1Pod result;
        String filePath = "/app/delete_pod_result.json";
        String apiUrl = String.format("%s/api/v1/namespaces/%s/pods/%s",myConfig.getApiServer(),namespace, podName);
        System.out.println(String.format("The constructed api url for deleting node is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X DELETE %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,myConfig.getToken(),filePath)
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
    private String getPodLog(String podName,String containerName){
        String log = "";

        String filePath = "/app/get_pod_log.json";
        String apiUrl = String.format("%s/api/v1/namespaces/%s/pods/%s/log?container=%s",myConfig.getApiServer(),NAMESPACE,podName,containerName);
        System.out.println(String.format("The constructed api url for getting the pod log is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,myConfig.getToken(),filePath)
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
    private boolean setServiceReplica(String serviceName, int targetNum){
        boolean status = false;
        SetServicesReplicasResponseFromAPI result;
        String filePath = "/app/set_service_replica.json";
        String apiUrl = String.format("%s/apis/extensions/v1beta1/namespaces/%s/deployments/%s/scale",myConfig.getApiServer() ,NAMESPACE,serviceName);
        System.out.println(String.format("The constructed api url is %s", apiUrl));
        String data ="'[{ \"op\": \"replace\", \"path\": \"/spec/replicas\", \"value\":" +  targetNum + " }]'";

        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X PATCH -d%s -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure >> %s",data,apiUrl,myConfig.getToken(),filePath)
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
    private boolean deleteNode(String nodeName){
        boolean isSuccess = true;
        DeleteNodeResult result;
        String filePath = "/app/delete_node_result.json";
        String apiUrl = String.format("%s/api/v1/nodes/%s",myConfig.getApiServer(),nodeName );
        System.out.println(String.format("The constructed api url for deleting node is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X DELETE %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,myConfig.getToken(),filePath)
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
    private void deleteService(String serviceName){
        DeleteServiceResult result;
        String filePath = "/app/delete_service_result.json";
        String apiUrl = String.format("%s/api/v1/namespaces/%s/services/%s",myConfig.getApiServer(), NAMESPACE,serviceName );
        System.out.println(String.format("The constructed api url for deleting service is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X DELETE %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,myConfig.getToken(),filePath)
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
    private boolean isAllReady(String namespace){
        boolean isAllReady = true;
        QueryDeploymentsListResponse deploymentsList = getDeploymentList(namespace);
        for(SingleDeploymentInfo singleDeploymentInfo : deploymentsList.getItems()){
            if(singleDeploymentInfo.getStatus().getReplicas() != singleDeploymentInfo.getStatus().getReadyReplicas()){
                isAllReady = false;
                break;
            }
        }
        return isAllReady;
    }

    //Check if all the required deployment replicas are ready
    private boolean isAllReady(SetServiceReplicasRequest setServiceReplicasRequest){
        boolean isAllReady = true;

        QueryDeploymentsListResponse deploymentsList = getDeploymentList(NAMESPACE);

        for(ServiceReplicasSetting setting : setServiceReplicasRequest.getServiceReplicasSettings()){
            if(!isSingleReady(deploymentsList.getItems(),setting)){
                isAllReady = false;
                break;
            }
        }
        return isAllReady;
    }

    //Get the deployment list
    private QueryDeploymentsListResponse getDeploymentList(String namespace){
        //Get the current deployments information and echo to the file
        String filePath = "/app/get_deployment_list_result.json";
        QueryDeploymentsListResponse deploymentsList = new QueryDeploymentsListResponse();
        String apiUrl = String.format("%s/apis/apps/v1beta1/namespaces/%s/deployments",myConfig.getApiServer() ,namespace);
        System.out.println(String.format("The constructed api url for getting the deploymentlist is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,myConfig.getToken(),filePath)
        };
        ProcessBuilder pb = new ProcessBuilder(cmds);
        pb.redirectErrorStream(true);
        Process p;
        try {
            p = pb.start();
            p.waitFor();

            //Wait 2 seconds to ensure the existence of the file
            Thread.sleep(2000);

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
    private V1NodeList getNodeList(){
        //Get the current deployments information and echo to the file
        String filePath = "/app/get_node_list_result.json";
        V1NodeList nodeList = new V1NodeList();
        String apiUrl = String.format("%s/api/v1/nodes",myConfig.getApiServer());
        System.out.println(String.format("The constructed api url for getting the node list is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,myConfig.getToken(),filePath)
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
    private V1PodList getPodList(String namespace){
        if(namespace.equals(""))
            namespace = NAMESPACE;
        //Get the current pods information and echo to the file
        String filePath = "/app/get_pod_list_result.json";
        V1PodList podList = new V1PodList();
        String apiUrl = String.format("%s/api/v1/namespaces/%s/pods",myConfig.getApiServer(),namespace);
        System.out.println(String.format("The constructed api url for getting the pod list is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,myConfig.getToken(),filePath)
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
    private V1Pod getPodInfo(String name){
        //Get the current pods information and echo to the file
        String filePath = "/app/get_pod_info_result.json";
        V1Pod pod = null;
        String apiUrl = String.format("%s/api/v1/namespaces/%s/pods/%s",myConfig.getApiServer(),NAMESPACE,name);
        System.out.println(String.format("The constructed api url for getting the pod info is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,myConfig.getToken(),filePath)
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
