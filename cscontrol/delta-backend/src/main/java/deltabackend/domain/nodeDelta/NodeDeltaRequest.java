package deltabackend.domain.nodeDelta;

import java.util.List;

public class NodeDeltaRequest {

    String id;
    List<String> nodeNames;
    String cluster = null;

    public String getCluster() {
        return cluster;
    }

    public void setCluster(String cluster) {
        this.cluster = cluster;
    }


    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public List<String> getNodeNames() {
        return nodeNames;
    }

    public void setNodeNames(List<String> nodeNames) {
        this.nodeNames = nodeNames;
    }

}
