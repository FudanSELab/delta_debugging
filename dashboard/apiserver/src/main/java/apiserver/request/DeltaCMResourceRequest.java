package apiserver.request;

import apiserver.bean.NewSingleDeltaCMResourceRequest;

import java.util.List;

public class DeltaCMResourceRequest {
    private String clusterName;
    private List<NewSingleDeltaCMResourceRequest> deltaRequests;

    public DeltaCMResourceRequest(){

    }

    public String getClusterName() {
        return clusterName;
    }

    public void setClusterName(String clusterName) {
        this.clusterName = clusterName;
    }

    public List<NewSingleDeltaCMResourceRequest> getDeltaRequests() {
        return deltaRequests;
    }

    public void setDeltaRequests(List<NewSingleDeltaCMResourceRequest> deltaRequests) {
        this.deltaRequests = deltaRequests;
    }
}
