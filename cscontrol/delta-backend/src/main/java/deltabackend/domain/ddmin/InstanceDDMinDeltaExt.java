package deltabackend.domain.ddmin;

import com.baeldung.algorithms.ddmin.DDMinDelta;
import deltabackend.domain.DeltaTestRequest;
import deltabackend.domain.DeltaTestResponse;
import deltabackend.domain.DeltaTestResult;
import deltabackend.domain.EnvParameter;
import deltabackend.domain.api.SetServiceReplicasRequest;
import deltabackend.domain.api.SetServiceReplicasResponse;
import deltabackend.domain.instanceDelta.DeltaResponse;
import org.apache.commons.collections4.CollectionUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Bean;
import org.springframework.messaging.MessageHeaders;
import org.springframework.messaging.SubscribableChannel;
import org.springframework.messaging.simp.SimpMessageHeaderAccessor;
import org.springframework.messaging.simp.SimpMessageType;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.util.*;


public class InstanceDDMinDeltaExt extends DDMinDelta {

    private RestTemplate restTemplate = new RestTemplate();

    private SimpMessagingTemplate template ;

    private String sessionId;

    private int instanceDeltaN = 3;

    private List<EnvParameter> orignalEnv;

    private Map<String, EnvParameter> deltaMap = new HashMap<String, EnvParameter>();

    private String expectException = "exception";


    public InstanceDDMinDeltaExt(List<String> tests, List<EnvParameter> env, String id, SimpMessagingTemplate t) {
        super();
        testcases = tests;
        orignalEnv = env;
        sessionId = id;
        deltas_all = new ArrayList<String>();
        template = t;
        for(EnvParameter p: env){
            EnvParameter q = new EnvParameter();
            q.setServiceName(p.getServiceName());
            q.setNumOfReplicas(instanceDeltaN);
            deltaMap.put(p.getServiceName(), q);
            deltas_all.add(p.getServiceName());
        }
        expectError = "fail";
        expectPass = "pass";
    }


    public boolean applyDelta(List<String> deltas) {
        // recovery to original cluster status
        SetServiceReplicasResponse ssrr1 = setInstanceNumOfServices(orignalEnv);
        if(! ssrr1.isStatus()){
           return false;
        }

        // apply delta
        List<EnvParameter> env = new ArrayList<EnvParameter>();
        for(String s: deltas){
            EnvParameter e = deltaMap.get(s);
            env.add(e);
        }
        SetServiceReplicasResponse ssrr2 = setInstanceNumOfServices(env);
        if( ! ssrr2.isStatus()){
            return false;
        }
        return true;
    }


    public String processAndGetResult(List<String> deltas, List<String> testcases) {
        // execute testcases
        try {
            Thread.sleep(30000);
        } catch (InterruptedException e) {
            e.printStackTrace();
            System.out.println();
        }

        DeltaTestResponse result = deltaTests(testcases);
        List<EnvParameter> env = new ArrayList<EnvParameter>();
        System.out.println();
        System.out.println("***** processAndGetResult *****   " + deltas);
        System.out.println();
        for(String s: deltas){
            EnvParameter e = deltaMap.get(s);
            env.add(e);
        }
        responseToUser(env, result);

        String returnResult = "";
        if(result.getStatus() == 1){
            returnResult = expectPass;
        } else if(result.getStatus() == 0){
            returnResult = expectError;
        } else {
            returnResult = expectException;
        }
        System.out.println("******** returnResult *******" + returnResult);
        return returnResult;
    }


    private DeltaTestResponse deltaTests(List<String> testNames){
        DeltaTestRequest dtr = new DeltaTestRequest();
        dtr.setTestNames(testNames);
        DeltaTestResponse result = restTemplate.postForObject(
                "http://test-backend:5001/testBackend/deltaTest",dtr,
                DeltaTestResponse.class);
        return result;
    }


    private SetServiceReplicasResponse setInstanceNumOfServices(List<EnvParameter> env) {
        SetServiceReplicasRequest ssrr = new SetServiceReplicasRequest();
        ssrr.setServiceReplicasSettings(env);
        System.out.println();
        for(EnvParameter e: env){
            System.out.println("--setInstanceNumOfServices--" + e.getServiceName() + ": " + e.getNumOfReplicas());
        }
        SetServiceReplicasResponse ssresult = restTemplate.postForObject(
                "http://api-server:18898/api/setReplicas",ssrr,
                SetServiceReplicasResponse.class);
        System.out.println("--setInstanceNumOfServices--" + ssresult.isStatus() + ": " + ssresult.getMessage());
        System.out.println();
        return ssresult;
    }

    //////////////////////////////////// send result to user ////////////////////////////////////////////////////
    private void responseToUser(List<EnvParameter> env, DeltaTestResponse result){
        DeltaResponse dr = new DeltaResponse();
        if(result.getStatus() == -1){ //the backend throw an exception, stop the delta test, maybe the testcase not exist
            dr.setStatus(false);
            dr.setMessage(result.getMessage());
            template.convertAndSendToUser(sessionId,"/topic/deltaresponse" ,dr, createHeaders(sessionId));
        }
        dr.setStatus(true);//just mean the test case has been executed
        dr.setEnv(env);
        dr.setMessage(result.getMessage());
        dr.setResult(result);
        template.convertAndSendToUser(sessionId,"/topic/deltaresponse" ,dr, createHeaders(sessionId));
    }

    private MessageHeaders createHeaders(String sessionId) {
        SimpMessageHeaderAccessor headerAccessor = SimpMessageHeaderAccessor.create(SimpMessageType.MESSAGE);
        headerAccessor.setSessionId(sessionId);
        headerAccessor.setLeaveMutable(true);
        return headerAccessor.getMessageHeaders();
    }


}
