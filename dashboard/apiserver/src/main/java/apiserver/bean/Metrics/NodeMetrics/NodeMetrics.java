package apiserver.bean.metrics.NodeMetrics;

import apiserver.bean.metrics.Common.V1beta1ItemsUsage;

public class NodeMetrics {
    private String nodeId;
    private V1beta1ItemsUsage usage;

    public String getNodeId() {
        return nodeId;
    }

    public void setNodeId(String nodeId) {
        this.nodeId = nodeId;
    }

    public V1beta1ItemsUsage getUsage() {
        return usage;
    }

    public void setUsage(V1beta1ItemsUsage usage) {
        this.usage = usage;
    }
}
