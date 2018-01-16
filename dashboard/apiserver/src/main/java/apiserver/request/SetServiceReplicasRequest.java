package apiserver.request;

import apiserver.bean.ServiceReplicasSetting;

import java.util.List;

public class SetServiceReplicasRequest {
    private List<ServiceReplicasSetting> serviceReplicasSettings;

    public SetServiceReplicasRequest(){

    }

    public List<ServiceReplicasSetting> getServiceReplicasSettings() {
        return serviceReplicasSettings;
    }

    public void setServiceReplicasSettings(List<ServiceReplicasSetting> serviceReplicasSettings) {
        this.serviceReplicasSettings = serviceReplicasSettings;
    }
}
