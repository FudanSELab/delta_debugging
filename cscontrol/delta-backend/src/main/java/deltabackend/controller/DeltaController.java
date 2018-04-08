package deltabackend.controller;

import deltabackend.domain.configDelta.ConfigDeltaRequest;
import deltabackend.domain.instanceDelta.DeltaRequest;
import deltabackend.domain.instanceDelta.SimpleInstanceRequest;
import deltabackend.domain.nodeDelta.DeltaNodeByListResponse;
import deltabackend.domain.nodeDelta.DeltaNodeRequest;
import deltabackend.domain.nodeDelta.NodeDeltaRequest;
import deltabackend.domain.sequenceDelta.SequenceDeltaRequest;
import deltabackend.domain.serviceDelta.ExtractServiceRequest;
import deltabackend.domain.serviceDelta.ReserveServiceResponse;
import deltabackend.domain.serviceDelta.ServiceDeltaRequest;
import deltabackend.service.DeltaService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.web.bind.annotation.*;

@RestController
public class DeltaController {

    @Autowired
    private DeltaService deltaService;

    @CrossOrigin(origins = "*")
    @RequestMapping(value="/welcome", method = RequestMethod.GET)
    public String welcome() {
        return "hello";
    }


    @MessageMapping("/msg/delta")
    public void delta(DeltaRequest message) throws Exception {
        deltaService.delta(message);
    }

    @MessageMapping("/msg/simpleSetInstance")
    public void simpleSetInstance(SimpleInstanceRequest message) throws Exception {
        deltaService.simpleSetInstance(message);
    }

    ////////////////////////////////////////Service delta////////////////////////////////////////
    @MessageMapping("/msg/serviceDelta")
    public void serviceDelta(ServiceDeltaRequest message) throws Exception {
        deltaService.serviceDelta(message);
    }

    ////////////////////////////////////////Service delta////////////////////////////////////////
    @MessageMapping("/msg/nodeDelta")
    public void nodeDelta(NodeDeltaRequest message) throws Exception {
        deltaService.nodeDelta(message);
    }

    @CrossOrigin(origins = "*")
    @RequestMapping(value="/delta/extractServices", method = RequestMethod.POST)
    public ReserveServiceResponse extractServices(@RequestBody ExtractServiceRequest request) {
        return deltaService.extractServices(request);
    }

    @CrossOrigin(origins = "*")
    @RequestMapping(value="/delta/deleteNodes", method = RequestMethod.POST)
    public DeltaNodeByListResponse deleteNodes(@RequestBody DeltaNodeRequest request) {
        return deltaService.deleteNodesByList(request);
    }

    //////////////////////////////////////////To do //////////////////////////////////////////

    @MessageMapping("/msg/configDelta")
    public void configDelta(ConfigDeltaRequest message) throws Exception {
        deltaService.configDelta(message);
    }

    @MessageMapping("/msg/sequenceDelta")
    public void sequenceDelta(SequenceDeltaRequest message) throws Exception {
        deltaService.sequenceDelta(message);
    }

}
