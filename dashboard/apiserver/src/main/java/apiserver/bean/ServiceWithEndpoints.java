package apiserver.bean;

import java.util.List;

public class ServiceWithEndpoints {
    private String serviceName;
    private List<String> endPoints;

    public ServiceWithEndpoints(){

    }

    public String getServiceName() {
        return serviceName;
    }

    public void setServiceName(String serviceName) {
        this.serviceName = serviceName;
    }

    public List<String> getEndPoints() {
        return endPoints;
    }

    public void setEndPoints(List<String> endPoints) {
        this.endPoints = endPoints;
    }
}
