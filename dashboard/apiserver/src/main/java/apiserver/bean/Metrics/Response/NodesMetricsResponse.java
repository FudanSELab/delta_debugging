package apiserver.bean.Metrics.Response;

import java.util.List;

public class NodesMetricsResponse {
    public List<NodeMetrics> getNodesMetrics() {
        return nodesMetrics;
    }

    public void setNodesMetrics(List<NodeMetrics> nodesMetrics) {
        this.nodesMetrics = nodesMetrics;
    }

    private List<NodeMetrics> nodesMetrics;
}
