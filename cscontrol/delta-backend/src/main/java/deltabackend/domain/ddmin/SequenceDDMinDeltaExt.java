package deltabackend.domain.ddmin;

import com.baeldung.algorithms.ddmin.ParallelDDMinDelta;


import deltabackend.domain.DeltaTestRequest;
import deltabackend.domain.DeltaTestResponse;

import deltabackend.domain.sequenceDelta.SequenceDeltaResponse;
import deltabackend.domain.sequenceDelta.SetAsyncRequestSequenceRequestWithSource;
import deltabackend.domain.sequenceDelta.SetAsyncRequestSequenceResponse;
import org.springframework.messaging.MessageHeaders;
import org.springframework.messaging.simp.SimpMessageHeaderAccessor;
import org.springframework.messaging.simp.SimpMessageType;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.web.client.RestTemplate;

import java.util.ArrayList;
import java.util.List;

public class SequenceDDMinDeltaExt extends ParallelDDMinDelta {

    private RestTemplate restTemplate = new RestTemplate();

    private SimpMessagingTemplate template ;

    private String sessionId;

    private String expectException = "exception";

    private String sender;

    private ArrayList<String> receivers;

    public SequenceDDMinDeltaExt (List<String> c, List<String> tests, String s, ArrayList<String> rs, String id, SimpMessagingTemplate t){
        super();
        expectError = "fail";
        expectPass = "pass";
        testcases = tests;
        sessionId = id;
        clusters = c;
        sender = s;
        receivers = rs;

        deltas_all = new ArrayList<String>();
        for(int i = 0; i < rs.size()-1; i++){
            for(int j = i + 1; j < rs.size(); j++){
                deltas_all.add(rs.get(j) +  ":" + rs.get(i));
            }
        }
        System.out.println("######### delta_all ###########  " + deltas_all);
    }

    public boolean applyDelta(List<String> deltas) {
        // recovery to original cluster status
         SetAsyncRequestSequenceResponse r1 = releaseControl();
        if(! r1.isStatus()){
            return false;
        }

        // apply delta
        ArrayList<String> env = deltaToServiceOrder(deltas);
        SetAsyncRequestSequenceResponse r2 = controlSequence(env);
        if( ! r2.isStatus()){
            return false;
        }
        return true;
    }

    //get the right order according to the deltas
    private ArrayList<String> deltaToServiceOrder(List<String> deltas){

    }

    private SetAsyncRequestSequenceResponse releaseControl() {
        SetAsyncRequestSequenceRequestWithSource sar = new SetAsyncRequestSequenceRequestWithSource();
        sar.setSourceName(sender);
        sar.setSvcList(receivers);
        System.out.println("------ Recovering ----");
        SetAsyncRequestSequenceResponse r = restTemplate.postForObject(
                "http://api-server:18898/api/xxxxxx",sar,
                 SetAsyncRequestSequenceResponse.class);
        System.out.println("----  Recovering  --" + r.isStatus() + ": " + r.getMessage());
        System.out.println();
        return r;
    }

    private SetAsyncRequestSequenceResponse controlSequence(ArrayList<String> env) {
        SetAsyncRequestSequenceRequestWithSource sar = new SetAsyncRequestSequenceRequestWithSource();
        sar.setSourceName(sender);
        sar.setSvcList(env);
        System.out.println("---- controlSequence ----" + env);
        SetAsyncRequestSequenceResponse r = restTemplate.postForObject(
                "http://api-server:18898/api/setAsyncRequestSequenceWithSrc",sar,
                SetAsyncRequestSequenceResponse.class);
        System.out.println("--controlSequence--" + r.isStatus() + ": " + r.getMessage());
        System.out.println();
        return r;
    }

    public String processAndGetResult(List<String> deltas, List<String> testcases) {
        // execute testcases
        DeltaTestResponse result = deltaTests(testcases);
//        List<SingleDeltaCMResourceRequest> env = new ArrayList<SingleDeltaCMResourceRequest>();
//        for(String s: deltas){
//            SingleDeltaCMResourceRequest e = deltaMap.get(s);
//            env.add(e);
//        }
        System.out.println();
        System.out.println("***** processAndGetResult ***** delta ******* " + deltas);
        ArrayList<String> env = deltaToServiceOrder(deltas);
        System.out.println("***** processAndGetResult ***** env ********  " + env);
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


    //////////////////////////////////// send result to user ////////////////////////////////////////////////////
    private void responseToUser(List<String> order, DeltaTestResponse result){
        SequenceDeltaResponse sdr = new SequenceDeltaResponse();
        if(result.getStatus() == -1){ //the backend throw an exception, stop the delta test, maybe the testcase not exist
            sdr.setStatus(false);
            sdr.setMessage(result.getMessage());
            template.convertAndSendToUser(sessionId,"/topic/sequenceDeltaResponse" ,sdr, createHeaders(sessionId));
        }
        sdr.setStatus(true);//just mean the test case has been executed
        sdr.setSender(sender);
        sdr.setReceiversInOrder(order);
        sdr.setMessage(result.getMessage());
        sdr.setResult(result);
        template.convertAndSendToUser(sessionId,"/topic/sequenceDeltaResponse" ,sdr, createHeaders(sessionId));
    }

    private MessageHeaders createHeaders(String sessionId) {
        SimpMessageHeaderAccessor headerAccessor = SimpMessageHeaderAccessor.create(SimpMessageType.MESSAGE);
        headerAccessor.setSessionId(sessionId);
        headerAccessor.setLeaveMutable(true);
        return headerAccessor.getMessageHeaders();
    }



}
