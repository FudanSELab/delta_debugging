package deltabackend.controller;

import deltabackend.domain.*;
import deltabackend.domain.api.*;
import deltabackend.domain.socket.SocketSessionRegistry;
import deltabackend.service.DeltaService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.messaging.MessageHeaders;
import org.springframework.messaging.handler.annotation.MessageMapping;
import org.springframework.messaging.simp.SimpMessageHeaderAccessor;
import org.springframework.messaging.simp.SimpMessageType;
import org.springframework.messaging.simp.SimpMessagingTemplate;
import org.springframework.web.bind.annotation.CrossOrigin;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestMethod;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.client.RestTemplate;

import java.util.*;

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


}
