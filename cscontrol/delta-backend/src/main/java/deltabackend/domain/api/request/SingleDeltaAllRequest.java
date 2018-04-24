package deltabackend.domain.api.request;

import deltabackend.domain.configDelta.CMConfig;

import java.util.ArrayList;
import java.util.List;

public class SingleDeltaAllRequest {
    private String serviceName;
    private List<CMConfig> configs = new ArrayList<CMConfig>();
    private int numOfReplicas;

    public SingleDeltaAllRequest(){

    }

    public SingleDeltaAllRequest(String sn, int num){
        this.serviceName = sn;
        this.numOfReplicas = num;
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

    public String toString(){
        return this.serviceName + ": { numOfReplicas: " + this.numOfReplicas + ", configs: " + this.configs.toString() + " }";
    }
}
