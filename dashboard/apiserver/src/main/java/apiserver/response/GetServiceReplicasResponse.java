package apiserver.response;

import apiserver.bean.ServiceWithReplicas;

import java.util.List;

public class GetServiceReplicasResponse {
    private boolean status;
    private String message;
    private List<ServiceWithReplicas> services;

    public GetServiceReplicasResponse(){

    }

    public boolean isStatus() {
        return status;
    }

    public void setStatus(boolean status) {
        this.status = status;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public List<ServiceWithReplicas> getServices() {
        return services;
    }

    public void setServices(List<ServiceWithReplicas> services) {
        this.services = services;
    }
}
