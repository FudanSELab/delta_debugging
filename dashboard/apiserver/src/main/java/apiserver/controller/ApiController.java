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

    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/hello", method= RequestMethod.GET)
    public String hello(){
        return "hello api-server";
    }

    //Get the available clusters
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getClusters", method= RequestMethod.GET)
    public GetClustersResponse getClusters(){
        return apiService.getClusters();
    }


    /**
     * 将所有发往某个服务的请求的传送延迟设置为10000秒（拦截请求）
     * (尽管传入参数中有source服务和destination服务，但是source服务这个参数并未被使用)
     */
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/setServiceRequestSuspend", method= RequestMethod.POST)
    public SetUnsetServiceRequestSuspendResponse setServiceRequestSuspend(
            @RequestBody SetUnsetServiceRequestSuspendRequest request){

        System.out.println("[=====] /api/setServiceRequestSuspend");
        System.out.println("[=====] svcName:" + request.getSvc());

        return apiService.setServiceRequestSuspend(request);
    }

    /**
     * 解除掉发往某个服务的传送延迟（请求拦截）
     */
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/unsetServiceRequestSuspend", method= RequestMethod.POST)
    public SetUnsetServiceRequestSuspendResponse unsetServiceRequestSuspend(
            @RequestBody SetUnsetServiceRequestSuspendRequest request){

        System.out.println("[=====] /api/unsetServiceRequestSuspend");
        System.out.println("[=====] svcName:" + request.getSvc());

        return apiService.unsetServiceRequestSuspend(request);
    }

    /**
     * 批量解除掉发往某组服务的传送延迟（批量解除请求拦截）
     */
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/unsuspendAllRequests", method = RequestMethod.POST)
    public SetAsyncRequestSequenceResponse unsuspendAllRequests(
            @RequestBody SetAsyncRequestSequenceRequestWithSource request){

        System.out.println("[=====] /api/unsuspendAllRequests");
        System.out.println("[=====] src Name:" + request.getSourceName());
        System.out.println("[=====] svc Name:" + request.getSvcList().size());

        return apiService.unsuspendAllRequest(request);
    }


    /**
     * 将两个服务之间的请求传送延迟设置为10000秒（拦截请求）
     */
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/setServiceRequestSuspendWithSourceSvc", method= RequestMethod.POST)
    public SetUnsetServiceRequestSuspendResponse setServiceRequestSuspendWithSourceSvc(
            @RequestBody SetUnsetServiceRequestSuspendRequest request){

        System.out.println("[=====] /api/setServiceRequestSuspendWithSourceSvc");
        System.out.println("[=====] svcName:" + request.getSvc());
        System.out.println("[=====] sourceSvcName:" + request.getSourceSvcName());

        return apiService.setServiceRequestSuspendWithSource(request);
    }


    /**
     * 按顺序依次解除发往一组服务的请求的顺序，变相达到控制请求顺序的目的
     */
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/setAsyncRequestSequence", method= RequestMethod.POST)
    public SetAsyncRequestSequenceResponse setAsyncRequestSequenceResponse(
            @RequestBody SetAsyncRequestSequenceRequest request){

        System.out.println("[=====] /api/setAsyncRequestSequence");
        System.out.println("[=====] svc Number:" + request.getSvcList().size());

        return apiService.setAsyncRequestsSequence(request);
    }

    /**
     * 按顺序依次解除一个服务发往一组服务之间的请求的顺序，变相达到控制请求顺序的目的
     */
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/setAsyncRequestSequenceWithSrc", method = RequestMethod.POST)
    public SetAsyncRequestSequenceResponse setAsyncRequestSequenceResponseWithSrc(
            @RequestBody SetAsyncRequestSequenceRequestWithSource request){

        System.out.println("[=====] /api/setAsyncRequestSequence");
        System.out.println("[=====] src Name:" + request.getSourceName());
        System.out.println("[=====] svc Name:" + request.getSvcList().size());

        return apiService.setAsyncRequestsSequenceWithSource(request);
    }

    /**
     * 先设置好一个服务发往一组服务的请求的延迟，然后调用一个异步线程依次放开请求
     */
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/setAsyncRequestSequenceWithSrcCombineWithFullSuspend", method = RequestMethod.POST)
    public SetAsyncRequestSequenceResponse setAsyncRequestSequenceWithSrcCombineWithFullSuspend(
            @RequestBody SetAsyncRequestSequenceRequestWithSource request){

        System.out.println("[=====] /api/setAsyncRequestSequenceWithSrcCombineWithFullSuspend");
        System.out.println("[=====] src Name:" + request.getSourceName());
        System.out.println("[=====] svc Name:" + request.getSvcList().size());

        return apiService.setAsyncRequestSequenceWithSrcCombineWithFullSuspend(request);
    }

    /**
     * 先设置好一个服务发往一组服务的请求的延迟，然后调用一个异步线程依次放开请求
     * 只不过在每次放开请求之后，延迟会被重新加上去，以便模拟持续不断的请求控制
     */
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/controlSequenceAndMaintainIt", method = RequestMethod.POST)
    public SetAsyncRequestSequenceResponse controlSequenceAndMaintainIt(
            @RequestBody SetAsyncRequestSequenceRequestWithSource request){

        System.out.println("[=====] /api/controlSequenceAndMaintainIt");
        System.out.println("[=====] Control From:" + request.getSourceName());
        System.out.println("[=====] Control To:" + request.getSvcList().toString());

        return apiService.controlSequenceAndMaintainIt(request);

    }


    //Set the replicas of running service
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/setReplicas", method= RequestMethod.POST)
    public SetServiceReplicasResponse setServiceReplica(@RequestBody SetServiceReplicasRequest setServiceReplicasRequest){
        return apiService.setServiceReplica(setServiceReplicasRequest);
    }

    //Get the list of all current services: include name and replicas
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getServicesList/{clusterName}", method= RequestMethod.GET)
    public GetServicesListResponse getServicesList(@PathVariable String clusterName){
        return apiService.getServicesList(clusterName);
    }

    //Get the service config: currently cpu and memory only
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getServicesAndConfig/{clusterName}", method= RequestMethod.GET)
    public GetServicesAndConfigResponse getServicesAndConfig(@PathVariable String clusterName){
        return apiService.getServicesAndConfig(clusterName);
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
    @RequestMapping(value="/api/runOnSingleNode/{clusterName}", method= RequestMethod.GET)
    public SetRunOnSingleNodeResponse setRunOnSingleNode(@PathVariable String clusterName){
        return apiService.setRunOnSingleNode(clusterName);
    }

    //Get the list of all current nodes info
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getNodesList/{clusterName}", method= RequestMethod.GET)
    public GetNodesListResponse getNodesList(@PathVariable String clusterName){
        return apiService.getNodesList(clusterName);
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
    @RequestMapping(value="/api/getPodsList/{clusterName}", method= RequestMethod.GET)
    public GetPodsListResponse getPodsList(@PathVariable String clusterName){
        return apiService.getPodsListAPI(clusterName);
    }

    //Get all of the log of current pods
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getPodsLog/{clusterName}", method= RequestMethod.GET)
    public GetPodsLogResponse getPodsLog(@PathVariable String clusterName){
        return apiService.getPodsLog(clusterName);
    }

    //Get the log of the specific pod
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getSinglePodLog", method= RequestMethod.POST)
    public GetSinglePodLogResponse getSinglePodLog(@RequestBody GetSinglePodLogRequest getSinglePodLogRequest){
        return apiService.getSinglePodLog(getSinglePodLogRequest);
    }

    //Restart the zipkin pod
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/restartService/{clusterName}", method= RequestMethod.GET)
    public RestartServiceResponse restartService(@PathVariable String clusterName){
        return apiService.restartService(clusterName);
    }

    //Config the container resource
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/deltaCMResource", method= RequestMethod.POST)
    public DeltaCMResourceResponse deltaCMResource(@RequestBody DeltaCMResourceRequest deltaCMResourceRequest){
        return apiService.deltaCMResource(deltaCMResourceRequest);
    }

    //Get the endpoints of all services
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getServiceWithEndpoints/{clusterName}", method= RequestMethod.GET)
    public ServiceWithEndpointsResponse getServiceWithEndpoints(@PathVariable String clusterName){
        return apiService.getServiceWithEndpoints(clusterName);
    }

    //Get the endpoints of specific services
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/getSpecificServiceWithEndpoints", method= RequestMethod.POST)
    public ServiceWithEndpointsResponse getSpecificServiceWithEndpoints(@RequestBody ReserveServiceRequest reserveServiceRequest) {
        return apiService.getSpecificServiceWithEndpoints(reserveServiceRequest);
    }

    //Delta all: currently instance and config
    @CrossOrigin(origins = "*")
    @RequestMapping(value="/api/deltaAll", method= RequestMethod.POST)
    public SimpleResponse deltaAll(@RequestBody DeltaAllRequest deltaAllRequest) {
        return apiService.deltaAll(deltaAllRequest);
    }
}
