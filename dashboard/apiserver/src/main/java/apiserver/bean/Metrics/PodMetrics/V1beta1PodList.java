package apiserver.bean.Metrics.PodMetrics;

import apiserver.bean.Metrics.NodeMetrics.V1beta1NodeMetadata;

import java.util.List;

public class V1beta1PodList {
    private String kind;
    private String apiVersion;
    private V1beta1NodeMetadata metadata;
    private List<V1beta1PodItem> items;

    public String getKind() {
        return kind;
    }

    public void setKind(String kind) {
        this.kind = kind;
    }

    public String getApiVersion() {
        return apiVersion;
    }

    public void setApiVersion(String apiVersion) {
        this.apiVersion = apiVersion;
    }

    public V1beta1NodeMetadata getMetadata() {
        return metadata;
    }

    public void setMetadata(V1beta1NodeMetadata metadata) {
        this.metadata = metadata;
    }

    public List<V1beta1PodItem> getItems() {
        return items;
    }

    public void setItems(List<V1beta1PodItem> items) {
        this.items = items;
    }
}
