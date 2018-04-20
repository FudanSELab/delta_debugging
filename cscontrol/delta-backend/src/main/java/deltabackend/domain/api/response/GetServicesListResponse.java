package deltabackend.domain.api.response;



import deltabackend.domain.bean.ServiceWithReplicas;

import java.util.List;

public class GetServicesListResponse {
    private boolean status;
    private String message;
    private List<ServiceWithReplicas> services;

    public GetServicesListResponse(){

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
