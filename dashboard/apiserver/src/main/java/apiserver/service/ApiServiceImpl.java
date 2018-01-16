package apiserver.service;

import apiserver.bean.QueryDeploymentsListResponse;
import apiserver.bean.ServiceReplicasSetting;
import apiserver.bean.SetServicesReplicasResponseFromAPI;
import apiserver.bean.SingleDeploymentInfo;
import apiserver.request.SetServiceReplicasRequest;
import apiserver.response.SetServiceReplicasResponse;
import apiserver.util.Const;
import com.alibaba.fastjson.JSON;
import org.springframework.stereotype.Service;

import java.io.*;
import java.util.List;

@Service
public class ApiServiceImpl implements ApiService {

    private final String NAMESPACE = "default";

    @Override
    public SetServiceReplicasResponse setServiceReplica(SetServiceReplicasRequest setServiceReplicasRequest) {
        SetServiceReplicasResponse response = new SetServiceReplicasResponse();
        //Set the desired number of service replicas
        for(ServiceReplicasSetting setting : setServiceReplicasRequest.getServiceReplicasSettings()){
            String apiUrl = String.format("%s/apis/extensions/v1beta1/namespaces/%s/deployments/%s/scale",Const.APISERVER ,NAMESPACE,setting.getServiceName());
            System.out.println(String.format("The constructed api url is %s", apiUrl));
            String data ="'[{ \"op\": \"replace\", \"path\": \"/spec/replicas\", \"value\":" +  setting.getNumOfReplicas() + " }]'";

            String[] cmds ={
                    "/bin/sh","-c",String.format("curl -X PATCH -d%s -H 'Content-Type: application/json-patch+json' %s --header \"Authorization: Bearer %s\" --insecure",data,apiUrl,Const.TOKEN)
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
                System.out.println(responseBuilder.toString());
                SetServicesReplicasResponseFromAPI result = JSON.parseObject(responseBuilder.toString(), SetServicesReplicasResponseFromAPI.class);
                System.out.println(result.getKind());
                br.close();
            } catch (IOException e) {
                response.setStatus(false);
                response.setMessage(String.format("Exception: %s", e.getStackTrace()));
                e.printStackTrace();
            }

//            HttpHeaders headers = new HttpHeaders();
//            headers.set("Authorization",String.format("Bearer %s", Const.TOKEN));
//            headers.set("Content-Type","application/json-patch+json");
//            String data ="{ \"op\": \"replace\", \"path\": \"/spec/replicas\", \"value\":" +  setting.getNumOfReplicas() + " }";
//            HttpEntity<String> entity = new HttpEntity<String>(data, headers);
//
//            HttpComponentsClientHttpRequestFactory requestFactory = new HttpComponentsClientHttpRequestFactory();
//            restTemplate.setRequestFactory(requestFactory);
//
//            ResponseEntity<String> result = restTemplate.exchange(apiUrl, HttpMethod.PATCH, entity, String.class);

//            SetServicesReplicasResponseFromAPI result = restTemplate.patchForObject(apiUrl,entity,SetServicesReplicasResponseFromAPI.class);
//            System.out.println(String.format("The request result is %s", result));
//            RestTemplate restTemplate = new RestTemplate();
//            ResponseEntity<String> response = restTemplate.patchForObject(apiUrl,entity, String.class);

            //Set the token to get authorization
//            HttpHeaders headers = new HttpHeaders();
//            headers.set("Authorization", String.format("Bearer %s", Const.TOKEN));
//
//            Map<String,Object> payload = new HashMap<String,Object>();
//            payload.put("op","replace");
//            payload.put("path","/spec/replicas");
//            payload.put("value",setting.getNumOfReplicas());
//            JSONObject jsonObj = JSONObject.fromObject(payload);

//            HttpEntity<JSONObject> entity = new HttpEntity<JSONObject>(jsonObj,headers);
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

    //Check if the required deployment replicas are ready
    private boolean isAllReady(SetServiceReplicasRequest setServiceReplicasRequest){
        boolean isAllReady = true;
        String filePath = "/app/get_deployment_list_result.json";
        //Get the current deployments information and echo to the file
        QueryDeploymentsListResponse deploymentsList = new QueryDeploymentsListResponse();
        String apiUrl = String.format("%s/apis/apps/v1beta1/namespaces/%s/deployments",Const.APISERVER ,NAMESPACE);
        System.out.println(String.format("The constructed api url for getting the deploymentlist is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure >> %s",apiUrl,Const.TOKEN,filePath)
        };
        ProcessBuilder pb = new ProcessBuilder(cmds);
        pb.redirectErrorStream(true);
        Process p;
        try {
            p = pb.start();
            p.waitFor();

            String json = readWholeFile(filePath);
            //Parse the response to the SetServicesReplicasResponseFromAPI Bean
            System.out.println(json);
            deploymentsList = JSON.parseObject(json,QueryDeploymentsListResponse.class);
            System.out.println(deploymentsList.getItems().get(0).getMetadata().getName());
        } catch (IOException e) {
            e.printStackTrace();
        }catch(InterruptedException e){
            e.printStackTrace();
        }

        for(ServiceReplicasSetting setting : setServiceReplicasRequest.getServiceReplicasSettings()){
            if(!isSingleReady(deploymentsList.getItems(),setting)){
                isAllReady = false;
                break;
            }
        }
        return isAllReady;
    }

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
                System.out.println(file.getName() + "is deleted");
            } else {
                System.out.println("Delete failed.");
            }
        } catch (Exception e) {
            System.out.println("Exception occured when delete file");
            e.printStackTrace();
        }
    }
}
