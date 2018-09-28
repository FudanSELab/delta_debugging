package apiserver.bean.metrics.PodMetrics;

import apiserver.bean.metrics.Common.V1beta1ItemsUsage;

public class PodMetrics {
    private String podId;
    private String nodeId;
    private V1beta1ItemsUsage usage;

    public String getPodId() {
        return podId;
    }

    public void setPodId(String podId) {
        this.podId = podId;
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
