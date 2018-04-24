package deltabackend.domain.api.request;

import java.util.List;

public class DeltaAllRequest {
    private String clusterName;
    private List<SingleDeltaAllRequest> deltaRequests;

    public DeltaAllRequest(){

    }

    public String getClusterName() {
        return clusterName;
    }

    public void setClusterName(String clusterName) {
        this.clusterName = clusterName;
    }

    public List<SingleDeltaAllRequest> getDeltaRequests() {
        return deltaRequests;
    }

    public void setDeltaRequests(List<SingleDeltaAllRequest> deltaRequests) {
        this.deltaRequests = deltaRequests;
    }
}
