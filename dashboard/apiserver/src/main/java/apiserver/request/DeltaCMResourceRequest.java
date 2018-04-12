package apiserver.request;

import apiserver.bean.SingleDeltaCMResourceRequest;

import java.util.List;

public class DeltaCMResourceRequest {
    private String clusterName;
    private List<SingleDeltaCMResourceRequest> deltaRequests;

    public DeltaCMResourceRequest(){

    }

    public String getClusterName() {
        return clusterName;
    }

    public void setClusterName(String clusterName) {
        this.clusterName = clusterName;
    }

    public List<SingleDeltaCMResourceRequest> getDeltaRequests() {
        return deltaRequests;
    }

    public void setDeltaRequests(List<SingleDeltaCMResourceRequest> deltaRequests) {
        this.deltaRequests = deltaRequests;
    }
}
