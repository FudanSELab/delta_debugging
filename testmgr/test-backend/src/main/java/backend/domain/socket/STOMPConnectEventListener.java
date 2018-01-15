package backend.domain.socket;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.ApplicationListener;
import org.springframework.messaging.simp.stomp.StompHeaderAccessor;
import org.springframework.web.socket.messaging.SessionConnectEvent;

public class STOMPConnectEventListener  implements ApplicationListener<SessionConnectEvent> {

    @Autowired
    SocketSessionRegistry webAgentSessionRegistry;

    @Override
    public void onApplicationEvent(SessionConnectEvent event) {
        StompHeaderAccessor sha = StompHeaderAccessor.wrap(event.getMessage());
        //login get from browser
        String agentId = sha.getNativeHeader("login").get(0);
        System.out.println("-----agentId= " + agentId);
        String sessionId = sha.getSessionId();
        System.out.println("-----sessionId= " + sessionId);
        webAgentSessionRegistry.registerSessionId(agentId,sessionId);
    }
}