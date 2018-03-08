package apiserver.response;

import apiserver.bean.NodeInfo;
import apiserver.bean.ServiceWithReplicas;
import apiserver.bean.V1Node;

import java.util.List;

public class GetNodesListResponse {
    private boolean status;
    private String message;
    private List<NodeInfo> nodes;

    public GetNodesListResponse(){

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

    public List<NodeInfo> getNodes() {
        return nodes;
    }

    public void setNodes(List<NodeInfo> nodes) {
        this.nodes = nodes;
    }
}
