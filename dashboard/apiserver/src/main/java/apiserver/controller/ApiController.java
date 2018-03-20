package apiserver.controller;

import apiserver.request.DeltaNodeRequest;
import apiserver.request.GetServiceReplicasRequest;
import apiserver.request.ReserveServiceRequest;
import apiserver.request.SetServiceReplicasRequest;
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
}
