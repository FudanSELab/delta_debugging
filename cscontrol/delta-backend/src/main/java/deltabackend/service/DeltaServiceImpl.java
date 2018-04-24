package deltabackend.service;

import com.baeldung.algorithms.ddmin.DDMinAlgorithm;
import com.baeldung.algorithms.ddmin.DDMinDelta;
import com.baeldung.algorithms.ddmin.ParallelDDMinAlgorithm;
import com.baeldung.algorithms.ddmin.ParallelDDMinDelta;


import deltabackend.domain.api.request.*;
import deltabackend.domain.api.response.*;
import deltabackend.domain.bean.ServiceReplicasSetting;
import deltabackend.domain.bean.ServiceWithReplicas;

import deltabackend.domain.bean.SingleDeltaCMResourceRequest;
import deltabackend.domain.configDelta.*;
import deltabackend.domain.ddmin.ConfigDDMinDeltaExt;
import deltabackend.domain.ddmin.InstanceDDMinDeltaExt;
import deltabackend.domain.ddmin.MixerDDMinDeltaExt;
import deltabackend.domain.ddmin.SequenceDDMinDeltaExt;

import deltabackend.domain.instanceDelta.DeltaRequest;
import deltabackend.domain.instanceDelta.InstanceDDMinResponse;
import deltabackend.domain.instanceDelta.SimpleInstanceRequest;
import deltabackend.domain.mixerDelta.MixerDDMinResponse;
import deltabackend.domain.mixerDelta.MixerDeltaRequest;
import deltabackend.domain.nodeDelta.NodeDeltaRequest;
import deltabackend.domain.sequenceDelta.SequenceDDMinResponse;
import deltabackend.domain.sequenceDelta.SequenceDeltaRequest;
import deltabackend.domain.serviceDelta.ExtractServiceRequest;
import deltabackend.domain.serviceDelta.ReserveServiceResponse;
import deltabackend.domain.serviceDelta.ServiceDeltaRequest;
import deltabackend.domain.socket.SocketSessionRegistry;
import deltabackend.domain.test.DeltaTestRequest;
import deltabackend.domain.test.DeltaTestResponse;
import deltabackend.domain.test.DeltaTestResult;
import deltabackend.util.MyConfig;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.messaging.MessageHeaders;
import org.springframework.messaging.simp.SimpMessageHeaderAccessor;
import org.springframework.messaging.simp.SimpMessageType;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.util.*;
import java.util.concurrent.ExecutionException;

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
    public void delta(DeltaRequest message) throws ExecutionException, InterruptedException {
        if ( ! webAgentSessionRegistry.getSessionIds(message.getId()).isEmpty()){
            System.out.println("=============Get one Instance Delta Request=============");
            String sessionId=webAgentSessionRegistry.getSessionIds(message.getId()).stream().findFirst().get();
            System.out.println("sessionid = " + sessionId);
            List<String> envStrings= message.getEnv();
            //get cluster name
            String cluster = "cluster1";
            if(null != message.getCluster()){
                cluster = message.getCluster();
            }
            //query for the env services' instance number
            GetServiceReplicasResponse gsrp = queryServicesReplicas(envStrings, cluster);
            List<ServiceWithReplicas> env = null;
            if(gsrp.isStatus()){
                env = gsrp.getServices();
            } else {
                System.out.println("################ cannot get service replica number ####################");
            }

            ParallelDDMinAlgorithm ddmin = new ParallelDDMinAlgorithm();
            ParallelDDMinDelta ddmin_delta = new InstanceDDMinDeltaExt(message.getTests(),env, sessionId, template, myConfig.getClusters());
            ddmin.setDdmin_delta(ddmin_delta);
            ddmin.initEnv();
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
            ((InstanceDDMinDeltaExt)ddmin_delta).recoverEnv();
        }
    }


//    private DeltaTestResponse deltaTests(List<String> testNames){
//        DeltaTestRequest dtr = new DeltaTestRequest();
//        dtr.setTestNames(testNames);
//        DeltaTestResponse result = restTemplate.postForObject(
//                "http://test-backend:5001/testBackend/deltaTest",dtr,
//                DeltaTestResponse.class);
//        return result;
//    }

    //query for the env services' instance number
    private GetServiceReplicasResponse queryServicesReplicas(List<String> envStrings, String cluster){
        GetServiceReplicasRequest gsrr = new GetServiceReplicasRequest();
        gsrr.setServices(envStrings);
        gsrr.setClusterName(cluster);
        GetServiceReplicasResponse gsrp = restTemplate.postForObject(
                "http://api-server:18898/api/getServicesReplicas",gsrr,
                GetServiceReplicasResponse.class);
        System.out.println("============= GetServiceReplicasResponse =============");
        System.out.println(gsrp.toString());
        return gsrp;
    }


    //adjust the instance number
    @Override
    public SetServiceReplicasResponse setInstanceNumOfServices(List<ServiceReplicasSetting> env, String cluster) {
        SetServiceReplicasRequest ssrr = new SetServiceReplicasRequest();
        ssrr.setServiceReplicasSettings(env);
        ssrr.setClusterName(cluster);
        SetServiceReplicasResponse ssresult = restTemplate.postForObject(
                "http://api-server:18898/api/setReplicas",ssrr,
                SetServiceReplicasResponse.class);
        return ssresult;
    }


    @Override
    public void simpleSetInstance(SimpleInstanceRequest message) {
        if ( ! webAgentSessionRegistry.getSessionIds(message.getId()).isEmpty()){
            String sessionId=webAgentSessionRegistry.getSessionIds(message.getId()).stream().findFirst().get();
            List<ServiceReplicasSetting> env = new ArrayList<ServiceReplicasSetting>();
            for(String service: message.getServices()){
                ServiceReplicasSetting ep = new ServiceReplicasSetting();
                ep.setServiceName(service);
                ep.setNumOfReplicas(message.getInstanceNum());
                env.add(ep);
            }
            boolean success = true;
            for(String c : myConfig.getClusters()){
                SetServiceReplicasResponse ssrr = setInstanceNumOfServices(env, c);
                if( ! ssrr.isStatus()){
                    success = false;
                }
            }

            if(success){
                template.convertAndSendToUser(sessionId,"/topic/simpleSetInstanceResult" ,"Success to set these services' replica", createHeaders(sessionId));
            } else {
                template.convertAndSendToUser(sessionId,"/topic/simpleSetInstanceResult" ,"Fail to set these services' replica", createHeaders(sessionId));
            }
        }
    }


//    private boolean judgeDiffer(DeltaTestResponse first, DeltaTestResponse dtr){
//        List<DeltaTestResult> l1 = first.getDeltaResults();
//        List<DeltaTestResult> l2 = dtr.getDeltaResults();
//        if(l1.size() == l2.size()){
//            for(int i = 0; i < l1.size(); i ++){
//                if( ! l1.get(i).getStatus().equals(l2.get(i).getStatus())){
//                    return true;
//                }
//            }
//        } else {
//            return true;
//        }
//        return false;
//    }

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
            //get cluster name
            String cluster = "cluster1";
            if(null != message.getCluster()){
                cluster = message.getCluster();
            }

            RestartServiceResponse restartResult = restartZipkin(cluster);
            if(restartResult.isStatus()){
                runTestCases(message.getTests(), cluster);
                List<String> servicesNames = getServicesFromZipkin(cluster);
                ReserveServiceResponse response = extract(servicesNames, cluster);
                template.convertAndSendToUser(sessionId,"/topic/serviceDeltaResponse" ,response, createHeaders(sessionId));
            } else {
                ReserveServiceResponse response = new ReserveServiceResponse();
                response.setStatus(false);
                response.setMessage("Failed to restart the zipkin!");
                template.convertAndSendToUser(sessionId,"/topic/serviceDeltaResponse" ,response, createHeaders(sessionId));
            }

        }
    }

//    @Override
//    public ReserveServiceResponse extractServices(ExtractServiceRequest testCases) {
//        runTestCases(testCases.getTests(), cluster);
//        List<String> servicesNames = getServicesFromZipkin();
//        return extract(servicesNames);
//    }

    //zero, restart the zipkin
    private RestartServiceResponse restartZipkin(String cluster){
        RestartServiceResponse result = restTemplate.getForObject(
                "http://api-server:18898/api/restartService/" + cluster,
                RestartServiceResponse.class);
        return result;
    }

    //first, run testcases
    private void runTestCases(List<String> testCaseNames, String cluster){
        DeltaTestRequest dtr = new DeltaTestRequest();
        dtr.setTestNames(testCaseNames);
        dtr.setCluster(cluster);
        DeltaTestResponse result = restTemplate.postForObject(
                "http://test-backend:5001/testBackend/deltaTest",dtr,
                DeltaTestResponse.class);
    }

    //second, get all the needed services' name from zipkin
    public List<String> getServicesFromZipkin(String cluster){
        System.out.println("====== Zipkin URL ====" + myConfig.getZipkinUrl().get(cluster));
        List result = restTemplate.getForObject(
                myConfig.getZipkinUrl().get(cluster) + "/api/v1/services", List.class);
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
    private ReserveServiceResponse extract(List<String> serviceNames, String cluster){
        ReserveServiceRequest rsr = new ReserveServiceRequest();
        rsr.setServices(serviceNames);
        rsr.setClusterName(cluster);
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
            //get cluster name
            String cluster = "cluster1";
            if(null != message.getCluster()){
                cluster = message.getCluster();
            }

            DeltaNodeRequest list = new DeltaNodeRequest();
            list.setNodeNames(message.getNodeNames());
            list.setClusterName(cluster);
            DeltaNodeByListResponse result = restTemplate.postForObject(
                    "http://api-server:18898/api/deleteNodeByList",list,
                    DeltaNodeByListResponse.class);
            template.convertAndSendToUser(sessionId,"/topic/nodeDeltaResponse" ,result, createHeaders(sessionId));

        }
    }

//    @Override
//    public DeltaNodeByListResponse deleteNodesByList(DeltaNodeRequest list) {
//        DeltaNodeByListResponse result = restTemplate.postForObject(
//                "http://api-server:18898/api/deleteNodeByList",list,
//                DeltaNodeByListResponse.class);
//        return result;
//    }


    ///////////////////////////////////////Config Delta/////////////////////////////////////////////
    @Override
    public void configDelta(ConfigDeltaRequest message) throws ExecutionException, InterruptedException {
        if ( ! webAgentSessionRegistry.getSessionIds(message.getId()).isEmpty()){
            System.out.println("=============Get one Config Delta Request=============");
            String sessionId = webAgentSessionRegistry.getSessionIds(message.getId()).stream().findFirst().get();
            System.out.println("sessionid = " + sessionId);

            ParallelDDMinAlgorithm ddmin = new ParallelDDMinAlgorithm();
            ParallelDDMinDelta ddmin_delta = new ConfigDDMinDeltaExt(message.getTests(),message.getConfigs(), sessionId, template, myConfig.getClusters());
            ddmin.setDdmin_delta(ddmin_delta);
            ddmin.initEnv();
            List<String> ddminResult = ddmin.ddmin(ddmin_delta.deltas_all);

            ConfigDDMinResponse r = new ConfigDDMinResponse();
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

            template.convertAndSendToUser(sessionId,"/topic/configDeltaEnd" ,r, createHeaders(sessionId));
            ((ConfigDDMinDeltaExt)ddmin_delta).recoverEnv();
        }
    }

    @Override
    public void simpleSetOrignal(ConfigDeltaRequest message) {
        if ( ! webAgentSessionRegistry.getSessionIds(message.getId()).isEmpty()){
            System.out.println("=============Get one Config simpleSetOrignal Request=============");
            String sessionId = webAgentSessionRegistry.getSessionIds(message.getId()).stream().findFirst().get();
            System.out.println("sessionid = " + sessionId);

            boolean success = true;
            for(String c : myConfig.getClusters()){
                DeltaCMResourceRequest dcr = new DeltaCMResourceRequest();
                dcr.setDeltaRequests(transformToNewConfigDS(message.getConfigs()));
                dcr.setClusterName(c);
                DeltaCMResourceResponse r = restTemplate.postForObject(
                        "http://api-server:18898/api/deltaCMResource",dcr,
                        DeltaCMResourceResponse.class);
                if( ! r.isStatus()){
                    success = false;
                }
            }

            if(success){
                template.convertAndSendToUser(sessionId,"/topic/simpleSetOrignalResult" ,"Success to set these configs", createHeaders(sessionId));
            } else {
                template.convertAndSendToUser(sessionId,"/topic/simpleSetOrignalResult" ,"Fail to set these configs", createHeaders(sessionId));
            }
        }
    }

    private List<NewSingleDeltaCMResourceRequest> transformToNewConfigDS(List<SingleDeltaCMResourceRequest> list){
        List<NewSingleDeltaCMResourceRequest> newList = new ArrayList<NewSingleDeltaCMResourceRequest>();
        Set<String> existService = new HashSet<String>();
        for(SingleDeltaCMResourceRequest l: list){
            if(existService.contains(l.getServiceName())){
                for(NewSingleDeltaCMResourceRequest d: newList){
                    if(d.getServiceName().equals(l.getServiceName())){
                        for(CMConfig cm : d.getConfigs()){
                            if(cm.getType().equals(l.getType())){
                                cm.addValues(new CM(l.getKey(), l.getValue()));
                            }
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
                newL.setConfigs(newConfig);
            }
        }
        System.out.println("++++++++++ transformToNewConfigDS ++++++++++++ " + newList);
        return newList;
    }


    //////////////////////////////////////// Sequence Delta /////////////////////////////////////////////////

    @Override
    public void sequenceDelta(SequenceDeltaRequest message) throws ExecutionException, InterruptedException {
        if ( ! webAgentSessionRegistry.getSessionIds(message.getId()).isEmpty()){
            System.out.println("=============Get one Sequence Delta Request=============");
            String sessionId=webAgentSessionRegistry.getSessionIds(message.getId()).stream().findFirst().get();
            System.out.println("sessionid = " + sessionId);

            ParallelDDMinAlgorithm ddmin = new ParallelDDMinAlgorithm();
            ParallelDDMinDelta ddmin_delta = new SequenceDDMinDeltaExt(message.getTests(),message.getSeqGroups(), sessionId, template, myConfig.getClusters());
            ddmin.setDdmin_delta(ddmin_delta);
            ddmin.initEnv();
            List<String> ddminResult = ddmin.ddmin(ddmin_delta.deltas_all);
            System.out.println("######## ddminResult: " + ddminResult);

            List<String> finalResult = ((SequenceDDMinDeltaExt)ddmin_delta).getFinalResult(ddminResult);

            SequenceDDMinResponse r = new SequenceDDMinResponse();
            if(null != ddminResult){
                r.setStatus(true);
                r.setMessage("Success");
                r.setDdminResult(finalResult);
            } else {
                r.setStatus(false);
                r.setMessage("Failed");
                r.setDdminResult(null);
            }

            template.convertAndSendToUser(sessionId,"/topic/sequenceDeltaEnd" ,r, createHeaders(sessionId));
            ((SequenceDDMinDeltaExt)ddmin_delta).recoverEnv();
        }

    }

    //////////////////////////// Mixer Delta //////////////////////////////////////
    @Override
    public void mixerDelta(MixerDeltaRequest message) throws ExecutionException, InterruptedException {
        if ( ! webAgentSessionRegistry.getSessionIds(message.getId()).isEmpty()){
            System.out.println("=============Get one Mixer Delta Request=============");
            String sessionId=webAgentSessionRegistry.getSessionIds(message.getId()).stream().findFirst().get();
            System.out.println("sessionid = " + sessionId);

            ParallelDDMinAlgorithm ddmin = new ParallelDDMinAlgorithm();
            ParallelDDMinDelta ddmin_delta = new MixerDDMinDeltaExt(message.getTests(),message.getSeqGroups(),message.getInstances(), message.getConfigs(), sessionId, template, myConfig.getClusters());
            ddmin.setDdmin_delta(ddmin_delta);
            ddmin.initEnv();
            List<String> ddminResult = ddmin.ddmin(ddmin_delta.deltas_all);
            System.out.println("######## ddminResult: " + ddminResult);

            Map<String, List<String>> finalResult = ((MixerDDMinDeltaExt)ddmin_delta).getFinalResult(ddminResult);

            MixerDDMinResponse r = new MixerDDMinResponse();
            if(null != ddminResult){
                r.setStatus(true);
                r.setMessage("Success");
                r.setDdminResult(finalResult);
            } else {
                r.setStatus(false);
                r.setMessage("Failed");
                r.setDdminResult(null);
            }

            template.convertAndSendToUser(sessionId,"/topic/mixerDeltaEnd" ,r, createHeaders(sessionId));
            ((MixerDDMinDeltaExt)ddmin_delta).recoverEnv();
        }

    }

}
