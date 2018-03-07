package apiserver.request;

import java.util.List;

public class ReserveServiceRequest {
    private List<String> services;

    public ReserveServiceRequest(){

    }

    public List<String> getServices() {
        return services;
    }

    public void setServices(List<String> services) {
        this.services = services;
    }
}
