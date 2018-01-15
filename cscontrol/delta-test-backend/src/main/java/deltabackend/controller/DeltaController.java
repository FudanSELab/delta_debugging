package deltabackend.controller;

import deltabackend.domain.*;
import deltabackend.domain.socket.SocketSessionRegistry;
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

    @CrossOrigin(origins = "*")
    @RequestMapping(value="/welcome", method = RequestMethod.GET)
    public String welcome() {
        return "hello";
    }

    /**session manager*/
    @Autowired
    SocketSessionRegistry webAgentSessionRegistry;
    /**send message template*/
    @Autowired
    private SimpMessagingTemplate template;
    @Autowired
    private RestTemplate restTemplate;

    @MessageMapping("/msg/deltatest")
    public void deltaTest(DeltaRequest message) throws Exception {
        if ( ! webAgentSessionRegistry.getSessionIds(message.getId()).isEmpty()){
            String sessionId=webAgentSessionRegistry.getSessionIds(message.getId()).stream().findFirst().get();
            List<String> envStrings= message.getEnv();
            //query for the env services' instance number

            List<EnvParameter> env = new ArrayList<EnvParameter>();
            for(String es: envStrings){
                EnvParameter ep = new EnvParameter();
                ep.setServiceName(es);
                ep.setInstanceNum(3);
                env.add(ep);
            }

            for(int i = 0; i < env.size() + 1; i++){
                DeltaResponse dr = new DeltaResponse();
                List<EnvParameter> env2 = new ArrayList<EnvParameter>(env.size());
                Iterator<EnvParameter> iterator = env.iterator();
                while(iterator.hasNext()){
                    env2.add((EnvParameter) iterator.next().clone());
                }
                if( i != 0 && i <= env.size()){
                    env2.get(i-1).setInstanceNum(1);
                }
                dr.setEnv(env2);
//                Thread.sleep(2000);

                DeltaTestRequest dtr = new DeltaTestRequest();
                dtr.setTestNames(message.getTests());
                DeltaTestResponse result = restTemplate.postForObject(
                        "http://test-backend:5001/testBackend/deltaTest",dtr,
                        DeltaTestResponse.class);
                dr.setStatus(true);//just mean the test case has been executed
                dr.setMessage(result.getMessage());
                dr.setResult(result);
                template.convertAndSendToUser(sessionId,"/topic/deltaresponse" ,dr, createHeaders(sessionId));
//                if( ! result.isStatus()){ //if failure, break the loop
//                    break;
//                }
            }
        }
    }


    private MessageHeaders createHeaders(String sessionId) {
        SimpMessageHeaderAccessor headerAccessor = SimpMessageHeaderAccessor.create(SimpMessageType.MESSAGE);
        headerAccessor.setSessionId(sessionId);
        headerAccessor.setLeaveMutable(true);
        return headerAccessor.getMessageHeaders();
    }


}
