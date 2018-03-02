package apiserver.service;

import apiserver.bean.*;
import apiserver.request.GetServiceReplicasRequest;
import apiserver.request.SetServiceReplicasRequest;
import apiserver.response.GetServiceReplicasResponse;
import apiserver.response.GetServicesListResponse;
import apiserver.response.SetServiceReplicasResponse;
import apiserver.util.MyConfig;
import com.alibaba.fastjson.JSON;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

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
            System.out.println(deploymentsList.getItems().get(0).getMetadata().getName());
        } catch (IOException e) {
            e.printStackTrace();
        }catch(InterruptedException e){
            e.printStackTrace();
        }
        return deploymentsList;
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

    //Read the whole file
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
