package apiserver.request;

import java.util.List;

public class DeltaNodeRequest {
    private String clusterName;
    private List<String> nodeNames;

    public  DeltaNodeRequest(){

    }

    public String getClusterName() {
        return clusterName;
    }

    public void setClusterName(String clusterName) {
        this.clusterName = clusterName;
    }

    public List<String> getNodeNames() {
        return nodeNames;
    }

    public void setNodeNames(List<String> nodeNames) {
        this.nodeNames = nodeNames;
    }
}
