package deltabackend.domain.configDelta;


import java.util.List;

public class DeltaCMResourceRequest {
    private List<SingleDeltaCMResourceRequest> deltaRequests;

    public DeltaCMResourceRequest(){

    }

    public List<SingleDeltaCMResourceRequest> getDeltaRequests() {
        return deltaRequests;
    }

    public void setDeltaRequests(List<SingleDeltaCMResourceRequest> deltaRequests) {
        this.deltaRequests = deltaRequests;
    }
}
