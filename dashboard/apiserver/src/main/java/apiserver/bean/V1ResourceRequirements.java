package apiserver.bean;

import java.util.Map;

public class V1ResourceRequirements {
    private Map<String, String> limits = null;
    private Map<String, String> requests = null;

    public V1ResourceRequirements(){

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
}
