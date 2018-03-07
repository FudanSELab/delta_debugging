package apiserver.util;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Component
@ConfigurationProperties(prefix="myConfig") //接收application.yml中的myConfig下面的属性
public class MyConfig {

    private String apiServer;
    private String token;

    public String getApiServer() {
        return apiServer;
    }

    public void setApiServer(String apiServer) {
        this.apiServer = apiServer;
    }

    public String getToken() {
        return token;
    }

    public void setToken(String token) {
        this.token = token;
    }

}
