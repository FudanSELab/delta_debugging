package deltabackend.domain.ddmin;

import com.baeldung.algorithms.ddmin.ParallelDDMinDelta;
import deltabackend.domain.api.request.SetServiceReplicasRequest;
import deltabackend.domain.api.response.SetServiceReplicasResponse;
import deltabackend.domain.bean.ServiceWithReplicas;
import deltabackend.domain.test.DeltaTestRequest;
import deltabackend.domain.test.DeltaTestResponse;

import deltabackend.domain.bean.ServiceReplicasSetting;
import deltabackend.domain.instanceDelta.DeltaResponse;
import deltabackend.util.MyConfig;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.messaging.MessageHeaders;
import org.springframework.messaging.simp.SimpMessageHeaderAccessor;
import org.springframework.messaging.simp.SimpMessageType;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.web.client.RestTemplate;

import java.util.*;


public class InstanceDDMinDeltaExt extends ParallelDDMinDelta {

    private RestTemplate restTemplate = new RestTemplate();

    private SimpMessagingTemplate template ;

    private String sessionId;

    private int instanceDeltaN = 2;

    private List<ServiceWithReplicas> orignalEnv;

    private Map<String, ServiceWithReplicas> deltaMap = new HashMap<String, ServiceWithReplicas>();

    private String expectException = "exception";

    public InstanceDDMinDeltaExt(List<String> tests, List<ServiceWithReplicas> env, String id, SimpMessagingTemplate t, List<String> cs) {
        super();
        clusters = cs;
        testcases = tests;
        orignalEnv = env;
        sessionId = id;
        deltas_all = new ArrayList<String>();
        template = t;
        for(ServiceWithReplicas p: env){
            ServiceWithReplicas q = new ServiceWithReplicas();
            q.setServiceName(p.getServiceName());
            q.setNumOfReplicas(instanceDeltaN);
            deltaMap.put(p.getServiceName(), q);
            deltas_all.add(p.getServiceName());
        }
        expectError = "fail";
        expectPass = "pass";
    }


    public boolean applyDelta(List<String> deltas, String cluster) {
        // recovery to original cluster status
//        SetServiceReplicasResponse ssrr1 = setInstanceNumOfServices(orignalEnv, cluster);
//        if(! ssrr1.isStatus()){
//           return false;
//        }

        // apply delta
        List<ServiceWithReplicas> env = new ArrayList<ServiceWithReplicas>();
        for(String s: deltas){
            ServiceWithReplicas e = deltaMap.get(s);
            env.add(e);
        }
        for(ServiceWithReplicas swr1: orignalEnv){
            boolean toAdjust = false;
            for(ServiceWithReplicas swr2: env){
                if(swr1.getServiceName().equals(swr2.getServiceName())){
                    toAdjust = true;
                }
            }
            if(toAdjust == false){
                env.add(swr1);
            }
        }
        System.out.println("&&&& instance deltas to apply: &&&&& " + env);
        SetServiceReplicasResponse ssrr2 = setInstanceNumOfServices(env, cluster);
        if( ! ssrr2.isStatus()){
            return false;
        }
        return true;
    }

    public boolean recoverEnv(){
        for(String s : clusters){
            SetServiceReplicasResponse ssrr1 = setInstanceNumOfServices(orignalEnv, s);
            if(! ssrr1.isStatus()){
                return false;
            }
        }
        return true;
    }


    public String processAndGetResult(List<String> deltas, List<String> testcases, String cluster) {
        // execute testcases
        try {
            Thread.sleep(20000);
        } catch (InterruptedException e) {
            e.printStackTrace();
            System.out.println();
        }

        DeltaTestResponse result = deltaTests(testcases, cluster);
        List<ServiceWithReplicas> env = new ArrayList<ServiceWithReplicas>();
        System.out.println();
        System.out.println("***** processAndGetResult *****   " + deltas);
        System.out.println();
        for(String s: deltas){
            ServiceWithReplicas e = deltaMap.get(s);
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


    private DeltaTestResponse deltaTests(List<String> testNames, String cluster){
        DeltaTestRequest dtr = new DeltaTestRequest();
        dtr.setTestNames(testNames);
        dtr.setCluster(cluster);
        DeltaTestResponse result = restTemplate.postForObject(
                "http://test-backend:5001/testBackend/deltaTest",dtr,
                DeltaTestResponse.class);
        return result;
    }


    private SetServiceReplicasResponse setInstanceNumOfServices(List<ServiceWithReplicas> env, String cluster) {
        SetServiceReplicasRequest ssrr = new SetServiceReplicasRequest();
        List<ServiceReplicasSetting> l = new ArrayList<ServiceReplicasSetting>();
        for(ServiceWithReplicas swr: env){
            ServiceReplicasSetting srs = new ServiceReplicasSetting();
            srs.setServiceName(swr.getServiceName());
            srs.setNumOfReplicas(swr.getNumOfReplicas());
            l.add(srs);
        }
        ssrr.setServiceReplicasSettings(l);
        ssrr.setClusterName(cluster);

        System.out.println();
        for(ServiceWithReplicas e: env){
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
    private void responseToUser(List<ServiceWithReplicas> env, DeltaTestResponse result){
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
