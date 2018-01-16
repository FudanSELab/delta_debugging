package apiserver.controller;

import apiserver.request.GetServiceReplicasRequest;
import apiserver.request.SetServiceReplicasRequest;
import apiserver.response.GetServiceReplicasResponse;
import apiserver.response.GetServicesListResponse;
import apiserver.response.SetServiceReplicasResponse;
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

    //Get the list of all current services
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
}
