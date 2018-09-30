package apiserver.bean.Metrics.NodeMetrics;

import apiserver.bean.Metrics.Common.V1beta1ItemsUsage;

public class NodeMetrics {
    private String nodeId;
    private V1beta1ItemsUsage usage;
    private V1beta1ItemsUsage config;

    public V1beta1ItemsUsage getConfig() {
        return config;
    }

    public void setConfig(V1beta1ItemsUsage config) {
        this.config = config;
    }

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
