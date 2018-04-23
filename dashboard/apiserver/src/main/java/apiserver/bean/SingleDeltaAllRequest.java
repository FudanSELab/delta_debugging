package apiserver.bean;

import java.util.List;

public class SingleDeltaAllRequest {
    private String serviceName;
    private List<CMConfig> configs;
    private int numOfReplicas;

    public SingleDeltaAllRequest(){

    }

    public String getServiceName() {
        return serviceName;
    }

    public void setServiceName(String serviceName) {
        this.serviceName = serviceName;
    }

    public List<CMConfig> getConfigs() {
        return configs;
    }

    public void setConfigs(List<CMConfig> configs) {
        this.configs = configs;
    }

    public int getNumOfReplicas() {
        return numOfReplicas;
    }

    public void setNumOfReplicas(int numOfReplicas) {
        this.numOfReplicas = numOfReplicas;
    }
}
