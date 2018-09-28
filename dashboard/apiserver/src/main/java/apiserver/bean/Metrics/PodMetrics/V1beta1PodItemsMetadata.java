package apiserver.bean.Metrics.PodMetrics;

import apiserver.bean.Metrics.NodeMetrics.V1beta1NodeItemsMetadata;

public class V1beta1PodItemsMetadata extends V1beta1NodeItemsMetadata {
    private String namespace;

    public String getNamespace() {
        return namespace;
    }

    public void setNamespace(String namespace) {
        this.namespace = namespace;
    }
}
