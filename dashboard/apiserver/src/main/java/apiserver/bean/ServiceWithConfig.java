package apiserver.bean;

import java.util.Map;

public class ServiceWithConfig {
    private String serviceName;
    private Map<String, String> limits;
    private Map<String, String> requests;
    private String instanceNumber;

    public ServiceWithConfig(){

    }

    public String getServiceName() {
        return serviceName;
    }

    public void setServiceName(String serviceName) {
        this.serviceName = serviceName;
    }

    public Map<String, String> getLimits() {
        return limits;
    }

    public void setLimits(Map<String, String> limits) {
        this.limits = limits;
    }

    public Map<String, String> getRequests() {
        return requests;
    }

    public void setRequests(Map<String, String> requests) {
        this.requests = requests;
    }

    public String  getInstanceNumber() {
        return instanceNumber;
    }

    public void setInstanceNumber(String  instanceNumber) {
        this.instanceNumber = instanceNumber;
    }
}
