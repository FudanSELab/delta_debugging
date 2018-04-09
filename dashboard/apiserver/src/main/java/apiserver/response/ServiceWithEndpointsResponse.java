package apiserver.response;

import apiserver.bean.ServiceWithEndpoints;

import java.util.List;

public class ServiceWithEndpointsResponse {
    private boolean status;
    private String message;
    private List<ServiceWithEndpoints> services;

    public ServiceWithEndpointsResponse(){

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

    public List<ServiceWithEndpoints> getServices() {
        return services;
    }

    public void setServices(List<ServiceWithEndpoints> services) {
        this.services = services;
    }
}
