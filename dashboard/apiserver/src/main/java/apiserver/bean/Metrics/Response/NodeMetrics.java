package apiserver.bean.Metrics.Response;

import apiserver.bean.Metrics.NodeMetrics.V1beta1NodeItemsUsage;

public class NodeMetrics {
    private String nodeId;
    private V1beta1NodeItemsUsage usage;

    public String getNodeId() {
        return nodeId;
    }

    public void setNodeId(String nodeId) {
        this.nodeId = nodeId;
    }

    public V1beta1NodeItemsUsage getUsage() {
        return usage;
    }

    public void setUsage(V1beta1NodeItemsUsage usage) {
        this.usage = usage;
    }
}
