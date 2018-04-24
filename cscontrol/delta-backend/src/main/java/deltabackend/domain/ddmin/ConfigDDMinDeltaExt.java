package deltabackend.domain.ddmin;

import com.baeldung.algorithms.ddmin.ParallelDDMinDelta;
import deltabackend.domain.api.request.DeltaCMResourceRequest;
import deltabackend.domain.api.response.DeltaCMResourceResponse;
import deltabackend.domain.bean.SingleDeltaCMResourceRequest;
import deltabackend.domain.configDelta.CM;
import deltabackend.domain.configDelta.CMConfig;
import deltabackend.domain.configDelta.ConfigDeltaResponse;
import deltabackend.domain.configDelta.NewSingleDeltaCMResourceRequest;
import deltabackend.domain.test.DeltaTestRequest;
import deltabackend.domain.test.DeltaTestResponse;


import org.springframework.messaging.MessageHeaders;
import org.springframework.messaging.simp.SimpMessageHeaderAccessor;
import org.springframework.messaging.simp.SimpMessageType;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.web.client.RestTemplate;

import java.util.*;


public class ConfigDDMinDeltaExt extends ParallelDDMinDelta {

    private RestTemplate restTemplate = new RestTemplate();

    private SimpMessagingTemplate template ;

    private String sessionId;

    private Map<String, String> unlimitMap = new HashMap<String, String>();

    private List<SingleDeltaCMResourceRequest> orignalEnv;

    private List<SingleDeltaCMResourceRequest> unlimitEnv;

    private Map<String, SingleDeltaCMResourceRequest> deltaMap = new HashMap<String, SingleDeltaCMResourceRequest>();

    private String expectException = "exception";


    public ConfigDDMinDeltaExt(List<String> tests, List<SingleDeltaCMResourceRequest> env, String id, SimpMessagingTemplate t, List<String> cs) {
        super();
        clusters = cs;
        unlimitMap.put("memory", "1000Mi");
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
        for(String s : clusters){
            DeltaCMResourceResponse r1 = modifyConfigsOfServices(transformToNewConfigDS(orignalEnv), s);
            if(! r1.isStatus()){
                return false;
            }
        }
        return true;
    }


    public boolean applyDelta(List<String> deltas, String cluster) {
        // recovery to original cluster status
//        DeltaCMResourceResponse r1 = modifyConfigsOfServices(unlimitEnv, cluster);
//        if(! r1.isStatus()){
//           return false;
//        }

        // apply delta
        List<SingleDeltaCMResourceRequest> env = new ArrayList<SingleDeltaCMResourceRequest>();
        for(String s: deltas){
            SingleDeltaCMResourceRequest e = deltaMap.get(s);
            env.add(e);
        }
        for(SingleDeltaCMResourceRequest sdcr1: unlimitEnv ){
            boolean toAdjust = false;
            for(SingleDeltaCMResourceRequest sdcr2: env){
                if(sdcr1.getServiceName().equals(sdcr2.getServiceName()) && sdcr1.getType().equals(sdcr2.getType()) && sdcr1.getKey().equals(sdcr2.getKey()) ){
                    toAdjust = true;
                }
            }
            if(toAdjust == false){
                env.add(sdcr1);
            }
        }
        DeltaCMResourceResponse r2 = modifyConfigsOfServices(transformToNewConfigDS(env), cluster);
        if( ! r2.isStatus()){
            return false;
        }
        return true;
    }

    private List<NewSingleDeltaCMResourceRequest> transformToNewConfigDS(List<SingleDeltaCMResourceRequest> list){
        System.out.println("^^^^ transformToNewConfigDS original ^^^^^^ " + list);

        List<NewSingleDeltaCMResourceRequest> newList = new ArrayList<NewSingleDeltaCMResourceRequest>();
        Set<String> existService = new HashSet<String>();
        for(SingleDeltaCMResourceRequest l: list){
            if(existService.contains(l.getServiceName())){
                for(NewSingleDeltaCMResourceRequest d: newList){
                    if(d.getServiceName().equals(l.getServiceName())){
                        int hasSameType = 0;
                        if( d.getConfigs().size() > 0){
                            for(CMConfig cm : d.getConfigs()){
                                if(cm.getType().equals(l.getType())){
                                    hasSameType = 1;
                                    cm.addValues(new CM(l.getKey(), l.getValue()));
                                    break;
                                }
                            }
                        }
                        if(hasSameType == 0){
                            CMConfig e = new CMConfig();
                            e.setType(l.getType());
                            e.addValues(new CM(l.getKey(), l.getValue()));
                            d.getConfigs().add(e);
                        }
                    }
                }
            } else {
                existService.add(l.getServiceName());
                NewSingleDeltaCMResourceRequest newL = new NewSingleDeltaCMResourceRequest();
                newL.setServiceName(l.getServiceName());
                List<CMConfig> newConfig = new ArrayList<CMConfig>();
                CMConfig cmc = new CMConfig();
                cmc.setType(l.getType());
                cmc.addValues(new CM(l.getKey(), l.getValue()));
                newConfig.add(cmc);
                newL.setConfigs(newConfig);
                newList.add(newL);
            }
        }
        System.out.println("++++++++++ transformToNewConfigDS ++++++++++++ " + newList);
        return newList;
    }


    public String processAndGetResult(List<String> deltas, List<String> testcases, String cluster) {
        // execute testcases
        DeltaTestResponse result = deltaTests(testcases, cluster);
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


    private DeltaTestResponse deltaTests(List<String> testNames, String cluster){
        DeltaTestRequest dtr = new DeltaTestRequest();
        dtr.setTestNames(testNames);
        dtr.setCluster(cluster);
        DeltaTestResponse result = restTemplate.postForObject(
                "http://test-backend:5001/testBackend/deltaTest",dtr,
                DeltaTestResponse.class);
        return result;
    }


    private DeltaCMResourceResponse modifyConfigsOfServices(List<NewSingleDeltaCMResourceRequest> env, String cluster) {
        DeltaCMResourceRequest dcr = new DeltaCMResourceRequest();
        dcr.setDeltaRequests(env);
        dcr.setClusterName(cluster);
        System.out.println();
        for(NewSingleDeltaCMResourceRequest e: env){
            System.out.println("--modifyConfigsOfServices--" + cluster + ": " + e.getServiceName() + ": " + e.getConfigs() );
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
