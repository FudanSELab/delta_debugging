package deltabackend.domain.ddmin;

import com.baeldung.algorithms.ddmin.ParallelDDMinDelta;


import deltabackend.domain.api.request.SetAsyncRequestSequenceRequestWithSource;
import deltabackend.domain.api.response.SetAsyncRequestSequenceResponse;
import deltabackend.domain.sequenceDelta.SequenceDeltaResponse;
import deltabackend.domain.sequenceDelta.SingleSequenceDelta;
import deltabackend.domain.test.DeltaTestRequest;
import deltabackend.domain.test.DeltaTestResponse;

import org.springframework.messaging.MessageHeaders;
import org.springframework.messaging.simp.SimpMessageHeaderAccessor;
import org.springframework.messaging.simp.SimpMessageType;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.web.client.RestTemplate;

import java.util.*;

public class SequenceDDMinDeltaExt extends ParallelDDMinDelta {

    private RestTemplate restTemplate = new RestTemplate();

    private SimpMessagingTemplate template ;

    private String sessionId;

    private String expectException = "exception";

    private Stack<String> seqNum = new Stack<String>();

    private Map<String, String> preToSender = new HashMap<String, String>();

    private Map<String, ArrayList<String>> preToReceivers = new HashMap<String, ArrayList<String>>();

    public SequenceDDMinDeltaExt (List<String> tests, List<SingleSequenceDelta> seqGroups, String id, SimpMessagingTemplate t, List<String> cs){
        super();
        seqNum.push("D");
        seqNum.push("C");
        seqNum.push("B");
        seqNum.push("A");

        clusters = cs;
        expectError = "fail";
        expectPass = "pass";
        testcases = tests;
        sessionId = id;
        template = t;
        clusters = cs;

        deltas_all = new ArrayList<String>();

        for(SingleSequenceDelta group: seqGroups){
            String s = group.getSender();
            List<String> rs = group.getReceivers();
            String prefix = "seq" + seqNum.pop();
            preToSender.put(prefix, s);
            ArrayList<String> l = new ArrayList<String>();
            //put inside-payment in the first
            for(String a : rs){
                if(a.contains("inside")){
                    l.add(a);
                }
            }
            for(String a : rs){
                if( ! a.contains("inside")){
                    l.add(a);
                }
            }
            System.out.println("------ New List -----" + l);
            preToReceivers.put(prefix, l);

            int size = rs.size();
            for(int i = 0; i < size-1; i++){
                for(int j = i + 1; j < size; j++){
                    deltas_all.add( prefix + "_" + size + "_" + (i+1) +  "_" + (j+1) );
                }
            }
        }

        System.out.println("######### delta_all ###########  " + deltas_all);
    }

    public boolean applyDelta(List<String> deltas, String cluster) {
        // recovery to original cluster status
//        if (recoverEnv()){
//            return false;
//        }

        // apply delta
        Map<String, String> seq_deltas = getSeqDeltas(deltas);

        List<String> seqUsed = new ArrayList<String>();
        if( ! seq_deltas.isEmpty()){
            for(String key: seq_deltas.keySet()){
                seqUsed.add(key);
                String sender = preToSender.get(key);
                String[] tmp = seq_deltas.get(key).split("_");
                ArrayList<String> receivers = preToReceivers.get(key);
                ArrayList<String> orders = new ArrayList<>();
                for(int i = 0; i < tmp.length; i++){
                    orders.add(receivers.get(Integer.parseInt(tmp[i]) - 1 ));
                }
                SetAsyncRequestSequenceResponse r2 = controlSequence(sender, orders, cluster);
                if( ! r2.isStatus()){
                    return false;
                }
            }
        }

        //if some seq not cover
        for(String key: preToSender.keySet()){
            if( ! seqUsed.contains(key)){
                String sender = preToSender.get(key);
                ArrayList<String> receivers = preToReceivers.get(key);
                SetAsyncRequestSequenceResponse r2 = controlSequence(sender, receivers, cluster);
                if(! r2.isStatus()){
                    return false;
                }
            }
        }

        return true;
    }


    public boolean recoverEnv(){
        for(String c: clusters){
            for(String key: preToSender.keySet()){
                String sender = preToSender.get(key);
                ArrayList<String> receivers = preToReceivers.get(key);
                SetAsyncRequestSequenceResponse r1 = releaseControl(sender, receivers, c);
                if(! r1.isStatus()){
                    return false;
                }
            }
        }
        return true;
    }


    private SetAsyncRequestSequenceResponse releaseControl(String sender, ArrayList<String> receivers, String cluster) {
        SetAsyncRequestSequenceRequestWithSource sar = new SetAsyncRequestSequenceRequestWithSource();
        sar.setSourceName(sender);
        sar.setSvcList(receivers);
        sar.setClusterName(cluster);
        System.out.println("------ Recovering ----");
        SetAsyncRequestSequenceResponse r = restTemplate.postForObject(
                "http://api-server:18898/api/unsuspendAllRequests",sar,
                SetAsyncRequestSequenceResponse.class);
        System.out.println("----  Recovering  --" + r.isStatus() + ": " + r.getMessage());
        System.out.println();

        return r;
    }

    private SetAsyncRequestSequenceResponse controlSequence(String sender, ArrayList<String> receivers, String cluster) {
        SetAsyncRequestSequenceRequestWithSource sar = new SetAsyncRequestSequenceRequestWithSource();
        sar.setSourceName(sender);
        sar.setSvcList(receivers);
        sar.setClusterName(cluster);
        System.out.println("---- controlSequence ----" + receivers);
        SetAsyncRequestSequenceResponse r = restTemplate.postForObject(
                "http://api-server:18898/api/setAsyncRequestSequenceWithSrcCombineWithFullSuspend",sar,
                SetAsyncRequestSequenceResponse.class);
        System.out.println("--controlSequence--" + r.isStatus() + ": " + r.getMessage());
        System.out.println();
        return r;
    }

    public List<String>  getFinalResult(List<String> ddminResult){
        List<String> r = new ArrayList<String>();
        if(null != ddminResult && ddminResult.size() > 0){
            for(String s : ddminResult){
                String[] tmp = s.split("_");
                String sender = preToSender.get(tmp[0]);
                String a = sender + "-> ";
                ArrayList<String> receivers = preToReceivers.get(tmp[0]);
                for(int i = 2; i < tmp.length; i++){
                    a += receivers.get(Integer.parseInt(tmp[i]) - 1) + " ";
                }
                a += "; ";
                r.add(a);
            }
        }
        return r;
    }

    private List<SetAsyncRequestSequenceRequestWithSource> transformToServiceNames(List<String> deltas, String cluster){
        //get full order
        Map<String, String> seq_deltas = getSeqDeltas(deltas);
        //get all the service names
        List<SetAsyncRequestSequenceRequestWithSource> envs = new ArrayList<SetAsyncRequestSequenceRequestWithSource>();
        for(String key: seq_deltas.keySet()){
            SetAsyncRequestSequenceRequestWithSource srsw = new SetAsyncRequestSequenceRequestWithSource();
            String sender = preToSender.get(key);
            srsw.setSourceName(sender);
            String[] tmp = seq_deltas.get(key).split("_");
            ArrayList<String> receivers = preToReceivers.get(key);
            ArrayList<String> rs = new ArrayList<String>();
            for(int i = 0; i < tmp.length; i++){
                rs.add(receivers.get(Integer.parseInt(tmp[i]) - 1 ));
            }
            srsw.setSvcList(rs);
            srsw.setClusterName(cluster);
            envs.add(srsw);
        }
        return envs;
    }

    public String processAndGetResult(List<String> deltas, List<String> testcases, String cluster) {
        // execute testcases
        try {
            Thread.sleep(15000);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }

        DeltaTestResponse result = deltaTests(testcases, cluster);

        System.out.println();
        System.out.println("***** processAndGetResult ***** delta ******* " + deltas);

//        Map<String, String> seq_deltas = getSeqDeltas(deltas);
//        List<SetAsyncRequestSequenceRequestWithSource> envs = new ArrayList<SetAsyncRequestSequenceRequestWithSource>();
//        for(String key: seq_deltas.keySet()){
//            SetAsyncRequestSequenceRequestWithSource srsw = new SetAsyncRequestSequenceRequestWithSource();
//            String sender = preToSender.get(key);
//            srsw.setSourceName(sender);
//            String[] tmp = seq_deltas.get(key).split("_");
//            ArrayList<String> receivers = preToReceivers.get(key);
//            ArrayList<String> rs = new ArrayList<String>();
//            for(int i = 0; i < tmp.length; i++){
//                rs.add(receivers.get(Integer.parseInt(tmp[i]) - 1 ));
//            }
//            srsw.setSvcList(rs);
//            srsw.setClusterName(cluster);
//            envs.add(srsw);
//        }

        responseToUser(transformToServiceNames(deltas, cluster), result);

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
        System.out.println("----- Get test response: " + result.getStatus() + result.getMessage());
        return result;
    }


    //////////////////////////////////// send result to user ////////////////////////////////////////////////////
    private void responseToUser(List<SetAsyncRequestSequenceRequestWithSource> envs, DeltaTestResponse result){
        SequenceDeltaResponse sdr = new SequenceDeltaResponse();
        if(result.getStatus() == -1){ //the backend throw an exception, stop the delta test, maybe the testcase not exist
            sdr.setStatus(false);
            sdr.setMessage(result.getMessage());
            template.convertAndSendToUser(sessionId,"/topic/sequenceDeltaResponse" ,sdr, createHeaders(sessionId));
        }
        sdr.setStatus(true);//just mean the test case has been executed
        sdr.setEnvList(envs);
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
