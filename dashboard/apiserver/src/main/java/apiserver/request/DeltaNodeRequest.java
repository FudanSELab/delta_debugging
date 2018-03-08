package apiserver.request;

import java.util.List;

public class DeltaNodeRequest {
    private List<String> nodeNames;

    public  DeltaNodeRequest(){

    }

    public List<String> getNodeNames() {
        return nodeNames;
    }

    public void setNodeNames(List<String> nodeNames) {
        this.nodeNames = nodeNames;
    }
}
