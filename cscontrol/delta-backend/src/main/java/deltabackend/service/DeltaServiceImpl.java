package deltabackend.service;

import deltabackend.domain.*;
import deltabackend.domain.api.GetServiceReplicasRequest;
import deltabackend.domain.api.GetServiceReplicasResponse;
import deltabackend.domain.api.SetServiceReplicasRequest;
import deltabackend.domain.api.SetServiceReplicasResponse;
import deltabackend.domain.socket.SocketSessionRegistry;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.messaging.MessageHeaders;
import org.springframework.messaging.simp.SimpMessageHeaderAccessor;
import org.springframework.messaging.simp.SimpMessageType;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.util.ArrayList;
import java.util.Iterator;
import java.util.List;

@Service
public class DeltaServiceImpl implements DeltaService{

    /**session manager*/
    @Autowired
    SocketSessionRegistry webAgentSessionRegistry;
    /**send message template*/
    @Autowired
    private SimpMessagingTemplate template;
    @Autowired
    private RestTemplate restTemplate;

    @Override
    public void delta(DeltaRequest message) {
        if ( ! webAgentSessionRegistry.getSessionIds(message.getId()).isEmpty()){
            System.out.println("=============Get one delta request=============");
            String sessionId=webAgentSessionRegistry.getSessionIds(message.getId()).stream().findFirst().get();
            System.out.println("sessionid = " + sessionId);
            List<String> envStrings= message.getEnv();
            //query for the env services' instance number
            GetServiceReplicasRequest gsrr = new GetServiceReplicasRequest();
            gsrr.setServices(envStrings);
            GetServiceReplicasResponse gsrp = restTemplate.postForObject(
                    "http://api-server:18898/api/getServicesReplicas",gsrr,
                    GetServiceReplicasResponse.class);
            System.out.println("============= GetServiceReplicasResponse =============");
            System.out.println(gsrp.toString());

            List<EnvParameter> env = null;
            if(gsrp.isStatus()){
                env = gsrp.getServices();
                System.out.println("============= GetServiceReplicasResponse status is true =============");
            } else {
                System.out.println("################ cannot get service replica number ####################");
            }

            DeltaTestResponse firstResult = new DeltaTestResponse();//save the first result
            for(int i = 0; null != env && i < env.size() + 1; i++){
                System.out.println("============= For loop to change the env parameter =============");
                if( i != 0 && env.get(i-1).getNumOfReplicas() <= 1){
                    continue;
                }
                DeltaResponse dr = new DeltaResponse();
                List<EnvParameter> env2 = new ArrayList<EnvParameter>(env.size());
                Iterator<EnvParameter> iterator = env.iterator();
                while(iterator.hasNext()){
                    env2.add((EnvParameter) iterator.next().clone());
                }
                if( i != 0 && i <= env.size()){
                    env2.get(i-1).setNumOfReplicas(1);
                }
                //adjust the instance number
                SetServiceReplicasRequest ssrr = new SetServiceReplicasRequest();
                ssrr.setServiceReplicasSettings(env2);
                SetServiceReplicasResponse ssresult = restTemplate.postForObject(
                        "http://api-server:18898/api/setReplicas",ssrr,
                        SetServiceReplicasResponse.class);
                if(ssresult.isStatus()){
                    System.out.println("============= SetServiceReplicasResponse status is true =============");
                    dr.setEnv(env2);
                    DeltaTestRequest dtr = new DeltaTestRequest();
                    dtr.setTestNames(message.getTests());
                    DeltaTestResponse result = restTemplate.postForObject(
                            "http://test-backend:5001/testBackend/deltaTest",dtr,
                            DeltaTestResponse.class);
                    dr.setStatus(true);//just mean the test case has been executed
                    dr.setMessage(result.getMessage());
                    dr.setResult(result);
                    if( i == 0 ){
                        firstResult = result;
                        dr.setDiffFromFirst(false);
                    } else {
                        dr.setDiffFromFirst(judgeDiffer( firstResult, result));
                    }
                    template.convertAndSendToUser(sessionId,"/topic/deltaresponse" ,dr, createHeaders(sessionId));
                } else {
                    System.out.println("-----------------" + ssresult.getMessage() + "----------------------");
                }
//                if( ! result.isStatus()){ //if failure, break the loop
//                    break;
//                }
            }
        }
    }

    private boolean judgeDiffer(DeltaTestResponse first, DeltaTestResponse dtr){
        List<DeltaTestResult> l1 = first.getDeltaResults();
        List<DeltaTestResult> l2 = dtr.getDeltaResults();
        if(l1.size() == l2.size()){
            for(int i = 0; i < l1.size(); i ++){
                if( ! l1.get(i).getStatus().equals(l2.get(i).getStatus())){
                    return true;
                }
            }
        } else {
            return true;
        }
        return false;
    }


    private MessageHeaders createHeaders(String sessionId) {
        SimpMessageHeaderAccessor headerAccessor = SimpMessageHeaderAccessor.create(SimpMessageType.MESSAGE);
        headerAccessor.setSessionId(sessionId);
        headerAccessor.setLeaveMutable(true);
        return headerAccessor.getMessageHeaders();
    }
}
