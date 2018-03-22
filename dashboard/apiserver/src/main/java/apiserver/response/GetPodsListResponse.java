package apiserver.response;

import apiserver.bean.NodeInfo;
import apiserver.bean.PodInfo;

import java.util.List;

public class GetPodsListResponse {
    private boolean status;
    private String message;
    private List<PodInfo> pods;

    public GetPodsListResponse(){

    }

    public boolean isStatus() {
        return status;
    }

    public void setStatus(boolean status) {
        this.status = status;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public List<PodInfo> getPods() {
        return pods;
    }

    public void setPods(List<PodInfo> pods) {
        this.pods = pods;
    }
}
