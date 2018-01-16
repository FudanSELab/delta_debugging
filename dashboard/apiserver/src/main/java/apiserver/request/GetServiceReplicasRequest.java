package apiserver.request;

import java.util.List;

public class GetServiceReplicasRequest {
    private List<String> services;

    public GetServiceReplicasRequest(){

    }

    public List<String> getServices() {
        return services;
    }

    public void setServices(List<String> services) {
        this.services = services;
    }
}
