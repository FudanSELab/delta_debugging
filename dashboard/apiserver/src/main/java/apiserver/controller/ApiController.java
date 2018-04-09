package apiserver.controller;

import apiserver.response.SetUnsetServiceRequestSuspendResponse;
import apiserver.request.*;
import apiserver.response.*;
import apiserver.service.ApiService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

@RestController
public class ApiController {

    @Autowired
    private ApiService apiService;

    //Set the replicas of running service

    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/hello", method= RequestMethod.GET)
    public String hello(){
        return "hello api-server";
    }

    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/setServiceRequestSuspend", method= RequestMethod.POST)
    public SetUnsetServiceRequestSuspendResponse setServiceRequestSuspend(
            @RequestBody SetUnsetServiceRequestSuspendRequest setUnsetServiceRequestSuspendRequest){

        System.out.println("[=====] /api/setServiceRequestSuspend");
        System.out.println("[=====] svcName:" + setUnsetServiceRequestSuspendRequest.getSvc());

        return apiService.setServiceRequestSuspend(setUnsetServiceRequestSuspendRequest);
    }

    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/setServiceRequestSuspendWithSourceSvc", method= RequestMethod.POST)
    public SetUnsetServiceRequestSuspendResponse setServiceRequestSuspendWithSourceSvc(
            @RequestBody SetUnsetServiceRequestSuspendRequest setUnsetServiceRequestSuspendRequest){

        System.out.println("[=====] /api/setServiceRequestSuspendWithSourceSvc");
        System.out.println("[=====] svcName:" + setUnsetServiceRequestSuspendRequest.getSvc());
        System.out.println("[=====] sourceSvcName:" + setUnsetServiceRequestSuspendRequest.getSourceSvcName());

        return apiService.setServiceRequestSuspendWithSource(setUnsetServiceRequestSuspendRequest);
    }


    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/unsetServiceRequestSuspend", method= RequestMethod.POST)
    public SetUnsetServiceRequestSuspendResponse unsetServiceRequestSuspend(
            @RequestBody SetUnsetServiceRequestSuspendRequest setUnsetServiceRequestSuspendRequest){

        System.out.println("[=====] /api/unsetServiceRequestSuspend");
        System.out.println("[=====] svcName:" + setUnsetServiceRequestSuspendRequest.getSvc());

        return apiService.unsetServiceRequestSuspend(setUnsetServiceRequestSuspendRequest);
    }

    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/setAsyncRequestSequence", method= RequestMethod.POST)
    public SetAsyncRequestSequenceResponse setAsyncRequestSequenceResponse(
            @RequestBody SetAsyncRequestSequenceRequest setAsyncRequestSequenceRequest){

        System.out.println("[=====] /api/setAsyncRequestSequence");
        System.out.println("[=====] svc Number:" + setAsyncRequestSequenceRequest.getSvcList().size());

        return apiService.setAsyncRequestsSequence(setAsyncRequestSequenceRequest);
    }

    //Set the replicas of running service
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/setReplicas", method= RequestMethod.POST)
    public SetServiceReplicasResponse setServiceReplica(@RequestBody SetServiceReplicasRequest setServiceReplicasRequest){
        return apiService.setServiceReplica(setServiceReplicasRequest);
    }

    //Get the list of all current services: include name and replicas
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getServicesList", method= RequestMethod.GET)
    public GetServicesListResponse getServicesList(){
        return apiService.getServicesList();
    }

    //Get the service config: currently cpu and memory only
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getServicesAndConfig", method= RequestMethod.GET)
    public GetServicesAndConfigResponse getServicesAndConfig(){
        return apiService.getServicesAndConfig();
    }

    //Get the number of replicas of services
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getServicesReplicas", method= RequestMethod.POST)
    public GetServiceReplicasResponse getServicesReplicas(@RequestBody GetServiceReplicasRequest getServiceReplicasRequest){
        return apiService.getServicesReplicas(getServiceReplicasRequest);
    }

    //Service delta: Reserve only the services contained in the list
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/reserveServiceByList", method= RequestMethod.POST)
    public ReserveServiceByListResponse reserveServiceByList(@RequestBody ReserveServiceRequest reserveServiceRequest){
        return apiService.reserveServiceByList(reserveServiceRequest);
    }

    //Node delta: Set to run on single node
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/runOnSingleNode", method= RequestMethod.GET)
    public SetRunOnSingleNodeResponse setRunOnSingleNode(){
        return apiService.setRunOnSingleNode();
    }

    //Get the list of all current nodes info
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getNodesList", method= RequestMethod.GET)
    public GetNodesListResponse getNodesList(){
        return apiService.getNodesList();
    }

    //Delete the node in the list
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/deleteNodeByList", method= RequestMethod.POST)
    public DeltaNodeByListResponse deleteNodeByList(@RequestBody DeltaNodeRequest deltaNodeRequest){
        return apiService.deleteNodeByList(deltaNodeRequest);
    }

    //Reserve the node in the list
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/reserveNodeByList", method= RequestMethod.POST)
    public DeltaNodeByListResponse reserveNodeByList(@RequestBody DeltaNodeRequest deltaNodeRequest){
        return apiService.reserveNodeByList(deltaNodeRequest);
    }

    //Get the list of all current pods info
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getPodsList", method= RequestMethod.GET)
    public GetPodsListResponse getPodsList(){
        return apiService.getPodsList();
    }

    //Get all of the log of current pods
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getPodsLog", method= RequestMethod.GET)
    public GetPodsLogResponse getPodsLog(){
        return apiService.getPodsLog();
    }

    //Get the log of the specific pod
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getSinglePodLog", method= RequestMethod.POST)
    public GetSinglePodLogResponse getSinglePodLog(@RequestBody GetSinglePodLogRequest getSinglePodLogRequest){
        return apiService.getSinglePodLog(getSinglePodLogRequest);
    }

    //Restart the zipkin pod
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/restartService", method= RequestMethod.GET)
    public RestartServiceResponse restartService(){
        return apiService.restartService();
    }

    //Config the container resource
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/deltaCMResource", method= RequestMethod.POST)
    public DeltaCMResourceResponse deltaCMResource(@RequestBody DeltaCMResourceRequest deltaCMResourceRequest){
        return apiService.deltaCMResource(deltaCMResourceRequest);
    }

    //Get the endpoints of all services
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getServiceWithEndpoints", method= RequestMethod.GET)
    public ServiceWithEndpointsResponse getServiceWithEndpoints(){
        return apiService.getServiceWithEndpoints();
    }

    //Get the endpoints of specific services
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getSpecificServiceWithEndpoints", method= RequestMethod.POST)
    public ServiceWithEndpointsResponse getSpecificServiceWithEndpoints(@RequestBody ReserveServiceRequest reserveServiceRequest) {
        return apiService.getSpecificServiceWithEndpoints(reserveServiceRequest);
    }
}
