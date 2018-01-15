package apiserver.service;

import apiserver.bean.ServiceReplicasSetting;
import apiserver.bean.SetServicesReplicasResponseFromAPI;
import apiserver.request.SetServiceReplicasRequest;
import apiserver.response.SetServiceReplicasResponse;
import apiserver.util.Const;
import net.sf.json.JSONObject;
import org.springframework.stereotype.Service;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;

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
                boolean record = false;
                StringBuilder responseBuilder = new StringBuilder();
                while((line = br.readLine()) != null){
                    if(line.contains("{"))
                        record = true;
                    if(record){
                        responseBuilder.append(line);
                    }
                }
                //Parse the response to the SetServicesReplicasResponseFromAPI Bean
                System.out.println(responseBuilder.toString());
                JSONObject obj = new JSONObject().fromObject(responseBuilder.toString());
                SetServicesReplicasResponseFromAPI result = (SetServicesReplicasResponseFromAPI)JSONObject.toBean(obj,SetServicesReplicasResponseFromAPI.class);
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
        if(isAllReady(setServiceReplicasRequest)){

        }
        return response;
    }

    //Check if the required deployment replicas are ready
    private boolean isAllReady(SetServiceReplicasRequest setServiceReplicasRequest){
        //Get the current deployments information
        String apiUrl = String.format("%s/apis/apps/v1beta1/namespaces/%s/deployments",Const.APISERVER ,NAMESPACE);
        System.out.println(String.format("The constructed api url is %s", apiUrl));
        String[] cmds ={
                "/bin/sh","-c",String.format("curl -X GET %s --header \"Authorization: Bearer %s\" --insecure",apiUrl,Const.TOKEN)
        };
        for(ServiceReplicasSetting setting : setServiceReplicasRequest.getServiceReplicasSettings()){
        }
        return false;
    }
}
