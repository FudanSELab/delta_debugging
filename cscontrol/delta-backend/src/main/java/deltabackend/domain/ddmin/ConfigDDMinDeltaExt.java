package deltabackend.domain.ddmin;

import com.baeldung.algorithms.ddmin.DDMinDelta;
import deltabackend.domain.DeltaTestRequest;
import deltabackend.domain.DeltaTestResponse;
import deltabackend.domain.EnvParameter;
import deltabackend.domain.api.SetServiceReplicasRequest;
import deltabackend.domain.api.SetServiceReplicasResponse;
import deltabackend.domain.configDelta.ConfigDeltaResponse;
import deltabackend.domain.configDelta.DeltaCMResourceRequest;
import deltabackend.domain.configDelta.DeltaCMResourceResponse;
import deltabackend.domain.configDelta.SingleDeltaCMResourceRequest;
import deltabackend.domain.instanceDelta.DeltaResponse;
import org.springframework.messaging.MessageHeaders;
import org.springframework.messaging.simp.SimpMessageHeaderAccessor;
import org.springframework.messaging.simp.SimpMessageType;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.web.client.RestTemplate;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;


public class ConfigDDMinDeltaExt extends DDMinDelta {

    private RestTemplate restTemplate = new RestTemplate();

    private SimpMessagingTemplate template ;

    private String sessionId;

    private Map<String, String> unlimitMap = new HashMap<String, String>();

    private List<SingleDeltaCMResourceRequest> orignalEnv;

    private List<SingleDeltaCMResourceRequest> unlimitEnv;

    private Map<String, SingleDeltaCMResourceRequest> deltaMap = new HashMap<String, SingleDeltaCMResourceRequest>();

    private String expectException = "exception";


    public ConfigDDMinDeltaExt(List<String> tests, List<SingleDeltaCMResourceRequest> env, String id, SimpMessagingTemplate t) {
        super();
        unlimitMap.put("memory", "800Mi");
        unlimitMap.put("cpu", "500m");
        unlimitEnv = new ArrayList<SingleDeltaCMResourceRequest>();
        for(SingleDeltaCMResourceRequest s : env){
            SingleDeltaCMResourceRequest a = new SingleDeltaCMResourceRequest();
            a.setServiceName(s.getServiceName());
            a.setType(s.getType());
            a.setKey(s.getKey());
            a.setValue(unlimitMap.get(s.getKey()));
            unlimitEnv.add(a);
        }

        testcases = tests;
        orignalEnv = env;
        sessionId = id;
        deltas_all = new ArrayList<String>();
        template = t;
        for(SingleDeltaCMResourceRequest p: env){
            SingleDeltaCMResourceRequest q = new SingleDeltaCMResourceRequest();
            q.setServiceName(p.getServiceName());
            q.setType(p.getType());
            q.setKey(p.getKey());
            q.setValue(p.getValue());
            deltaMap.put(q.getServiceName() + ":" + q.getType()+ ":" + q.getKey()+ ":" + q.getValue(), q);
            deltas_all.add(q.getServiceName() + ":" + q.getType()+ ":" + q.getKey()+ ":" + q.getValue());
        }
        expectError = "fail";
        expectPass = "pass";
    }


    public boolean recoverEnv(){
        DeltaCMResourceResponse r1 = modifyConfigsOfServices(orignalEnv);
        if(! r1.isStatus()){
            return false;
        }
        return true;
    }


    public boolean applyDelta(List<String> deltas) {
        // recovery to original cluster status
        DeltaCMResourceResponse r1 = modifyConfigsOfServices(unlimitEnv);
        if(! r1.isStatus()){
           return false;
        }

        // apply delta
        List<SingleDeltaCMResourceRequest> env = new ArrayList<SingleDeltaCMResourceRequest>();
        for(String s: deltas){
            SingleDeltaCMResourceRequest e = deltaMap.get(s);
            env.add(e);
        }
        DeltaCMResourceResponse r2 = modifyConfigsOfServices(env);
        if( ! r2.isStatus()){
            return false;
        }
        return true;
    }


    public String processAndGetResult(List<String> deltas, List<String> testcases) {
        // execute testcases
//        try {
//            Thread.sleep(120000);
//        } catch (InterruptedException e) {
//            e.printStackTrace();
//            System.out.println();
//        }

        DeltaTestResponse result = deltaTests(testcases);
        List<SingleDeltaCMResourceRequest> env = new ArrayList<SingleDeltaCMResourceRequest>();
        System.out.println();
        System.out.println("***** processAndGetResult *****   " + deltas);
        System.out.println();
        for(String s: deltas){
            SingleDeltaCMResourceRequest e = deltaMap.get(s);
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


    private DeltaCMResourceResponse modifyConfigsOfServices(List<SingleDeltaCMResourceRequest> env) {
        DeltaCMResourceRequest dcr = new DeltaCMResourceRequest();
        dcr.setDeltaRequests(env);
        System.out.println();
        for(SingleDeltaCMResourceRequest e: env){
            System.out.println("--modifyConfigsOfServices--" + e.getServiceName() + ": " + e.getType() + ": " + e.getKey() + ": " + e.getValue());
        }
        DeltaCMResourceResponse r = restTemplate.postForObject(
                "http://api-server:18898/api/deltaCMResource",dcr,
                DeltaCMResourceResponse.class);
        System.out.println("--modifyConfigsOfServices--" + r.isStatus() + ": " + r.getMessage());
        System.out.println();
        return r;
    }

    //////////////////////////////////// send result to user ////////////////////////////////////////////////////
    private void responseToUser(List<SingleDeltaCMResourceRequest> env, DeltaTestResponse result){
        ConfigDeltaResponse dr = new ConfigDeltaResponse();
        if(result.getStatus() == -1){ //the backend throw an exception, stop the delta test, maybe the testcase not exist
            dr.setStatus(false);
            dr.setMessage(result.getMessage());
            template.convertAndSendToUser(sessionId,"/topic/configDeltaResponse" ,dr, createHeaders(sessionId));
        }
        dr.setStatus(true);//just mean the test case has been executed
        dr.setEnv(env);
        dr.setMessage(result.getMessage());
        dr.setResult(result);
        template.convertAndSendToUser(sessionId,"/topic/configDeltaResponse" ,dr, createHeaders(sessionId));
    }

    private MessageHeaders createHeaders(String sessionId) {
        SimpMessageHeaderAccessor headerAccessor = SimpMessageHeaderAccessor.create(SimpMessageType.MESSAGE);
        headerAccessor.setSessionId(sessionId);
        headerAccessor.setLeaveMutable(true);
        return headerAccessor.getMessageHeaders();
    }


}
