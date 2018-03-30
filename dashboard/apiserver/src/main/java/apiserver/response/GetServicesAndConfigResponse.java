package apiserver.response;

import apiserver.bean.ServiceWithConfig;

import java.util.List;

public class GetServicesAndConfigResponse {
    private boolean status;
    private String message;
    private List<ServiceWithConfig> services;

    public GetServicesAndConfigResponse(){

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

    public List<ServiceWithConfig> getServices() {
        return services;
    }

    public void setServices(List<ServiceWithConfig> services) {
        this.services = services;
    }
}
