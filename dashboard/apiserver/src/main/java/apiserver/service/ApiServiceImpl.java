package apiserver.service;

import apiserver.bean.*;
import apiserver.request.GetServiceReplicasRequest;
import apiserver.request.ReserveServiceRequest;
import apiserver.request.SetServiceReplicasRequest;
import apiserver.response.*;
import apiserver.util.MyConfig;
import com.alibaba.fastjson.JSON;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import javax.print.attribute.standard.MediaSize;
import java.io.*;
import java.util.ArrayList;
import java.util.List;

@Service
public class ApiServiceImpl implements ApiService {

    private final String NAMESPACE = "default";
    @Autowired
    private MyConfig myConfig;


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
        QueryDeploymentsListResponse deploymentsList = getDeploymentList();
        //Iterate the list and return the result
        List<String> serviceNames = new ArrayList<String>();
        for(SingleDeploymentInfo singleDeploymentInfo : deploymentsList.getItems()){
            serviceNames.add(singleDeploymentInfo.getMetadata().getName());
        }
        System.out.println(String.format("The size of current service is %d",serviceNames.size()));
        if(deploymentsList.getItems().size() != 0){
            response.setServices(serviceNames);
            response.setMessage("Get the service name list successfully!");
            response.setStatus(true);
        }
        else{
            response.setStatus(false);
            response.setMessage("Fail to get the service name list!");
        }
        return response;
    }

    //Get the replicas num of the specific services
    @Override
    public GetServiceReplicasResponse getServicesReplicas(GetServiceReplicasRequest getServiceReplicasRequest) {
        GetServiceReplicasResponse response = new GetServiceReplicasResponse();
        //Get the current deployments information
        QueryDeploymentsListResponse deploymentsList = getDeploymentList();
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
            try{
                Thread.sleep(10000);
            }catch(Exception e){
                e.printStackTrace();
            }
            //Check whether all of the services are available again in certain internals
            while(!isAllReady()){
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

    //Reserve the services included in the list and delete the others
    @Override
    public ReserveServiceByListResponse reserveServiceByList(ReserveServiceRequest reserveServiceRequest) {
        ReserveServiceByListResponse response = new ReserveServiceByListResponse();
        response.setStatus(true);
        response.setMessage("Succeed to delete all of the services not contained in the list");
        //Get the current deployments information
        QueryDeploymentsListResponse deploymentsList = getDeploymentList();

        for(SingleDeploymentInfo singleDeploymentInfo : deploymentsList.getItems()){
            //Delete the services not contained in the list
            String deploymentName = singleDeploymentInfo.getMetadata().getName();
            if(!existInTheList(deploymentName,reserveServiceRequest.getServices())){
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
    private void deleteNode(String nodeName){
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
                System.out.println(String.format("Fail to delete the node %s. The corresponding status is %s",nodeName,result.getStatus()));
            }
        } catch (IOException e) {
            e.printStackTrace();
        }catch(InterruptedException e){
            e.printStackTrace();
        }
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
    private boolean isAllReady(){
        boolean isAllReady = true;
        QueryDeploymentsListResponse deploymentsList = getDeploymentList();
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

        QueryDeploymentsListResponse deploymentsList = getDeploymentList();

        for(ServiceReplicasSetting setting : setServiceReplicasRequest.getServiceReplicasSettings()){
            if(!isSingleReady(deploymentsList.getItems(),setting)){
                isAllReady = false;
                break;
            }
        }
        return isAllReady;
    }

    //Get the deployment list
    private QueryDeploymentsListResponse getDeploymentList(){
        //Get the current deployments information and echo to the file
        String filePath = "/app/get_deployment_list_result.json";
        QueryDeploymentsListResponse deploymentsList = new QueryDeploymentsListResponse();
        String apiUrl = String.format("%s/apis/apps/v1beta1/namespaces/%s/deployments",myConfig.getApiServer() ,NAMESPACE);
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

            String json = readWholeFile(filePath);
            //Parse the response to the SetServicesReplicasResponseFromAPI Bean
//            System.out.println(json);
            deploymentsList = JSON.parseObject(json,QueryDeploymentsListResponse.class);
        } catch (IOException e) {
            e.printStackTrace();
        }catch(InterruptedException e){
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
