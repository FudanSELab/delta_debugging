package apiserver.controller;

import apiserver.request.SetServiceReplicasRequest;
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
    public SetServiceReplicasResponse create(@RequestBody SetServiceReplicasRequest setServiceReplicasRequest){
        return apiService.setServiceReplica(setServiceReplicasRequest);
    }
}
