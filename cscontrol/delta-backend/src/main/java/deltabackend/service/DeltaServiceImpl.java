package deltabackend.service;

import com.baeldung.algorithms.ddmin.DDMinAlgorithm;
import com.baeldung.algorithms.ddmin.DDMinDelta;
import deltabackend.domain.*;
import deltabackend.domain.api.GetServiceReplicasRequest;
import deltabackend.domain.api.GetServiceReplicasResponse;
import deltabackend.domain.api.SetServiceReplicasRequest;
import deltabackend.domain.api.SetServiceReplicasResponse;
import deltabackend.domain.configDelta.ConfigDeltaRequest;
import deltabackend.domain.ddmin.InstanceDDMinDeltaExt;
import deltabackend.domain.ddmin.InstanceDDMinResponse;
import deltabackend.domain.instanceDelta.DeltaRequest;
import deltabackend.domain.instanceDelta.DeltaResponse;
import deltabackend.domain.instanceDelta.SimpleInstanceRequest;
import deltabackend.domain.nodeDelta.DeltaNodeByListResponse;
import deltabackend.domain.nodeDelta.DeltaNodeRequest;
import deltabackend.domain.nodeDelta.NodeDeltaRequest;
import deltabackend.domain.sequenceDelta.SequenceDeltaRequest;
import deltabackend.domain.serviceDelta.*;
import deltabackend.domain.socket.SocketSessionRegistry;
import deltabackend.util.MyConfig;
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

    @Autowired
    private MyConfig myConfig;


    //////////////////////////////////////////Instance Delta////////////////////////////////////////////////////
    @Override
    public void delta(DeltaRequest message) {
        if ( ! webAgentSessionRegistry.getSessionIds(message.getId()).isEmpty()){
            System.out.println("=============Get one delta request=============");
            String sessionId=webAgentSessionRegistry.getSessionIds(message.getId()).stream().findFirst().get();
            System.out.println("sessionid = " + sessionId);
            List<String> envStrings= message.getEnv();
            //query for the env services' instance number
            GetServiceReplicasResponse gsrp = queryServicesReplicas(envStrings);
            List<EnvParameter> env = null;
            if(gsrp.isStatus()){
                env = gsrp.getServices();
            } else {
                System.out.println("################ cannot get service replica number ####################");
            }

            DDMinAlgorithm ddmin = new DDMinAlgorithm();
            DDMinDelta ddmin_delta = new InstanceDDMinDeltaExt(message.getTests(),env, sessionId, template);
            ddmin.setDdmin_delta(ddmin_delta);
            List<String> ddminResult = ddmin.ddmin(ddmin_delta.deltas_all);

            InstanceDDMinResponse r = new InstanceDDMinResponse();
            if(null != ddminResult){
                for(String s: ddminResult){
                    System.out.println("######## ddminResult: " + s);
                    r.setStatus(true);
                    r.setMessage("Success");
                    r.setDdminResult(ddminResult);
                }
            } else {
                r.setStatus(false);
                r.setMessage("Failed");
                r.setDdminResult(null);
            }

            template.convertAndSendToUser(sessionId,"/topic/deltaEnd" ,r, createHeaders(sessionId));
        }
    }

//    @Override
//    public void delta(DeltaRequest message) {
//        if ( ! webAgentSessionRegistry.getSessionIds(message.getId()).isEmpty()){
//            System.out.println("=============Get one delta request=============");
//            String sessionId=webAgentSessionRegistry.getSessionIds(message.getId()).stream().findFirst().get();
//            System.out.println("sessionid = " + sessionId);
//            List<String> envStrings= message.getEnv();
//            //query for the env services' instance number
//            GetServiceReplicasResponse gsrp = queryServicesReplicas(envStrings);
//
//            List<EnvParameter> env = null;
//            if(gsrp.isStatus()){
//                env = gsrp.getServices();
//            } else {
//                System.out.println("################ cannot get service replica number ####################");
//            }
//
//            DeltaTestResponse firstResult = new DeltaTestResponse();//save the first result
//            for(int i = 0; null != env && i < env.size() + 1; i++){
//                System.out.println("============= For loop to change the env parameter =============");
//                if( i != 0 && env.get(i-1).getNumOfReplicas() <= 1){
//                    continue;
//                }
//                DeltaResponse dr = new DeltaResponse();
//                List<EnvParameter> env2 = new ArrayList<EnvParameter>(env.size());
//                Iterator<EnvParameter> iterator = env.iterator();
//                while(iterator.hasNext()){
//                    env2.add((EnvParameter) iterator.next().clone());
//                }
//                if( i != 0 && i <= env.size()){
//                    env2.get(i-1).setNumOfReplicas(1);
//                }
//                //adjust the instance number
//                SetServiceReplicasResponse ssresult = setInstanceNumOfServices(env2);
//
//                if(ssresult.isStatus()){
//                    System.out.println("============= SetServiceReplicasResponse status is true =============");
//                    dr.setEnv(env2);
//
//                    //delta tests
//                    DeltaTestResponse result = deltaTests(message.getTests());
//
//                    if(result.getStatus() == -1){ //the backend throw an exception, stop the delta test, maybe the testcase not exist
//                        dr.setStatus(false);
//                        dr.setMessage(result.getMessage());
//                        template.convertAndSendToUser(sessionId,"/topic/deltaresponse" ,dr, createHeaders(sessionId));
//                        break;
//                    }
//                    dr.setStatus(true);//just mean the test case has been executed
//                    dr.setMessage(result.getMessage());
//                    dr.setResult(result);
//                    if( i == 0 ){
//                        firstResult = result;
//                        dr.setDiffFromFirst(false);
//                    } else {
//                        dr.setDiffFromFirst(judgeDiffer( firstResult, result));
//                    }
//                    template.convertAndSendToUser(sessionId,"/topic/deltaresponse" ,dr, createHeaders(sessionId));
//                } else {
//                    System.out.println("-----------------" + ssresult.getMessage() + "----------------------");
//                }
////                if( ! result.isStatus()){ //if failure, break the loop
////                    break;
////                }
//            }
//
//        }
//    }


    private DeltaTestResponse deltaTests(List<String> testNames){
        DeltaTestRequest dtr = new DeltaTestRequest();
        dtr.setTestNames(testNames);
        DeltaTestResponse result = restTemplate.postForObject(
                "http://test-backend:5001/testBackend/deltaTest",dtr,
                DeltaTestResponse.class);
        return result;
    }


    //query for the env services' instance number
    private GetServiceReplicasResponse queryServicesReplicas(List<String> envStrings){
        GetServiceReplicasRequest gsrr = new GetServiceReplicasRequest();
        gsrr.setServices(envStrings);
        GetServiceReplicasResponse gsrp = restTemplate.postForObject(
                "http://api-server:18898/api/getServicesReplicas",gsrr,
                GetServiceReplicasResponse.class);
        System.out.println("============= GetServiceReplicasResponse =============");
        System.out.println(gsrp.toString());
        return gsrp;
    }


    //adjust the instance number
    @Override
    public SetServiceReplicasResponse setInstanceNumOfServices(List<EnvParameter> env) {
        SetServiceReplicasRequest ssrr = new SetServiceReplicasRequest();
        ssrr.setServiceReplicasSettings(env);
        SetServiceReplicasResponse ssresult = restTemplate.postForObject(
                "http://api-server:18898/api/setReplicas",ssrr,
                SetServiceReplicasResponse.class);
        return ssresult;
    }


    @Override
    public void simpleSetInstance(SimpleInstanceRequest message) {
        if ( ! webAgentSessionRegistry.getSessionIds(message.getId()).isEmpty()){
            String sessionId=webAgentSessionRegistry.getSessionIds(message.getId()).stream().findFirst().get();
            List<EnvParameter> env = new ArrayList<EnvParameter>();
            for(String service: message.getServices()){
                EnvParameter ep = new EnvParameter();
                ep.setServiceName(service);
                ep.setNumOfReplicas(message.getInstanceNum());
                env.add(ep);
            }
            SetServiceReplicasResponse ssrr = setInstanceNumOfServices(env);
            if(ssrr.isStatus()){
                template.convertAndSendToUser(sessionId,"/topic/simpleSetInstanceResult" ,"Success to set these services' replica", createHeaders(sessionId));
            } else {
                template.convertAndSendToUser(sessionId,"/topic/simpleSetInstanceResult" ,"Fail to set these services' replica", createHeaders(sessionId));
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

    /////////////////////////////Service Delta/////////////////////////////////////

    @Override
    public void serviceDelta(ServiceDeltaRequest message) {
        if ( ! webAgentSessionRegistry.getSessionIds(message.getId()).isEmpty()){
            System.out.println("=============Get one service delta request=============");
            String sessionId=webAgentSessionRegistry.getSessionIds(message.getId()).stream().findFirst().get();
            System.out.println("sessionid = " + sessionId);
            RestartServiceResponse restartResult = restartZipkin();
            if(restartResult.isStatus()){
                runTestCases(message.getTests());
                List<String> servicesNames = getServicesFromZipkin();
                ReserveServiceResponse response = extract(servicesNames);
                template.convertAndSendToUser(sessionId,"/topic/serviceDeltaResponse" ,response, createHeaders(sessionId));
            } else {
                ReserveServiceResponse response = new ReserveServiceResponse();
                response.setStatus(false);
                response.setMessage("Failed to restart the zipkin!");
                template.convertAndSendToUser(sessionId,"/topic/serviceDeltaResponse" ,response, createHeaders(sessionId));
            }

        }
    }

    @Override
    public ReserveServiceResponse extractServices(ExtractServiceRequest testCases) {
        runTestCases(testCases.getTests());
        List<String> servicesNames = getServicesFromZipkin();
        return extract(servicesNames);
    }

    //zero, restart the zipkin
    private RestartServiceResponse restartZipkin(){
        RestartServiceResponse result = restTemplate.getForObject(
                "http://test-backend:5001/testBackend/deltaTest",
                RestartServiceResponse.class);
        return result;
    }

    //first, run testcases
    private void runTestCases(List<String> testCaseNames){
        DeltaTestRequest dtr = new DeltaTestRequest();
        dtr.setTestNames(testCaseNames);
        DeltaTestResponse result = restTemplate.postForObject(
                "http://test-backend:5001/testBackend/deltaTest",dtr,
                DeltaTestResponse.class);
    }

    //second, get all the needed services' name from zipkin
    public List<String> getServicesFromZipkin(){
        List result = restTemplate.getForObject(
                myConfig.getZipkinUrl() + "/api/v1/services", List.class);
        Iterator it = result.iterator();
        List<String> serviceNames = new ArrayList<String>();
        while(it.hasNext()){
            //cast to String
            String s = (String)it.next();
            serviceNames.add(s);
            System.out.println("======zipkin===services==name====");
            System.out.println("======  "+s+"  =====");
        }
        return serviceNames;
    }

    //third, let k8s stop the services except services names from zipkin
    private ReserveServiceResponse extract(List<String> serviceNames){
        ReserveServiceRequest rsr = new ReserveServiceRequest();
        rsr.setServices(serviceNames);
        ReserveServiceByListResponse result = restTemplate.postForObject(
                "http://api-server:18898/api/reserveServiceByList",rsr,
                ReserveServiceByListResponse.class);
        ReserveServiceResponse r = new ReserveServiceResponse();
        r.setStatus(result.isStatus());
        r.setMessage(result.getMessage());
        r.setServiceNames(serviceNames);
        return r;
    }


////////////////////////////////Node Delta/////////////////////////////////////////////
    @Override
    public void nodeDelta(NodeDeltaRequest message) {
        if ( ! webAgentSessionRegistry.getSessionIds(message.getId()).isEmpty()){
            System.out.println("=============Get one node delta request=============");
            String sessionId=webAgentSessionRegistry.getSessionIds(message.getId()).stream().findFirst().get();
            System.out.println("sessionid = " + sessionId);

            DeltaNodeRequest list = new DeltaNodeRequest();
            list.setNodeNames(message.getNodeNames());
            DeltaNodeByListResponse result = restTemplate.postForObject(
                    "http://api-server:18898/api/deleteNodeByList",list,
                    DeltaNodeByListResponse.class);
            template.convertAndSendToUser(sessionId,"/topic/nodeDeltaResponse" ,result, createHeaders(sessionId));

        }
    }

    @Override
    public DeltaNodeByListResponse deleteNodesByList(DeltaNodeRequest list) {
        DeltaNodeByListResponse result = restTemplate.postForObject(
                "http://api-server:18898/api/deleteNodeByList",list,
                DeltaNodeByListResponse.class);
        return result;
    }


    ///////////////////////////////////////To do///////////////////////////////////////////////////
    ///////////////////////////////////////Config Delta/////////////////////////////////////////////
    @Override
    public void configDelta(ConfigDeltaRequest message) {
        if ( ! webAgentSessionRegistry.getSessionIds(message.getId()).isEmpty()){

        }
    }


    ////////////////////////////////////////Sequence Delta/////////////////////////////////////////////////

    @Override
    public void sequenceDelta(SequenceDeltaRequest message) {

    }

}
