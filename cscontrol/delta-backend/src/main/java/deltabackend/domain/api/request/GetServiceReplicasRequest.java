package deltabackend.domain.api.request;

import java.util.List;

public class GetServiceReplicasRequest {
    private String clusterName;
    private List<String> services;

    public GetServiceReplicasRequest(){

    }

    public String getClusterName() {
        return clusterName;
    }

    public void setClusterName(String clusterName) {
        this.clusterName = clusterName;
    }

    public List<String> getServices() {
        return services;
    }

    public void setServices(List<String> services) {
        this.services = services;
    }
}
