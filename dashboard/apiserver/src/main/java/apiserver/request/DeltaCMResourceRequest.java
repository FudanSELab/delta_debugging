package apiserver.request;

import apiserver.bean.SingleDeltaCMResourceRequest;

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
