package apiserver.bean.Metrics.Response;

import apiserver.bean.Metrics.NodeMetrics.NodeMetrics;

import java.util.List;

public class NodesMetricsResponse {
    private boolean status;
    private String message;
    private List<NodeMetrics> nodesMetrics;

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

    public List<NodeMetrics> getNodesMetrics() {
        return nodesMetrics;
    }

    public void setNodesMetrics(List<NodeMetrics> nodesMetrics) {
        this.nodesMetrics = nodesMetrics;
    }


}
