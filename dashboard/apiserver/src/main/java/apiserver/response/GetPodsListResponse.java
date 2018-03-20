package apiserver.response;

import apiserver.bean.NodeInfo;
import apiserver.bean.PodInfo;

import java.util.List;

public class GetPodsListResponse {
    private boolean status;
    private String message;
    private List<PodInfo> nodes;

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

    public List<PodInfo> getNodes() {
        return nodes;
    }

    public void setNodes(List<PodInfo> nodes) {
        this.nodes = nodes;
    }
}
