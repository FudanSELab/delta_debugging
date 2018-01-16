package deltabackend.domain.api;

import deltabackend.domain.EnvParameter;

import java.util.List;

public class SetServiceReplicasRequest {
    private List<EnvParameter> serviceReplicasSettings;

    public SetServiceReplicasRequest(){

    }

    public List<EnvParameter> getServiceReplicasSettings() {
        return serviceReplicasSettings;
    }

    public void setServiceReplicasSettings(List<EnvParameter> serviceReplicasSettings) {
        this.serviceReplicasSettings = serviceReplicasSettings;
    }
}
