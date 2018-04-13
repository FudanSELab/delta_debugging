package apiserver.request;

import apiserver.bean.ServiceReplicasSetting;

import java.util.List;

public class SetServiceReplicasRequest {
    private String clusterName;
    private List<ServiceReplicasSetting> serviceReplicasSettings;

    public SetServiceReplicasRequest(){

    }

    public String getClusterName() {
        return clusterName;
    }

    public void setClusterName(String clusterName) {
        this.clusterName = clusterName;
    }

    public List<ServiceReplicasSetting> getServiceReplicasSettings() {
        return serviceReplicasSettings;
    }

    public void setServiceReplicasSettings(List<ServiceReplicasSetting> serviceReplicasSettings) {
        this.serviceReplicasSettings = serviceReplicasSettings;
    }
}
