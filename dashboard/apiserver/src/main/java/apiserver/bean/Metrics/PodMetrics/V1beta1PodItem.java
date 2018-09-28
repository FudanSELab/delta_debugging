package apiserver.bean.metrics.PodMetrics;

import java.util.ArrayList;
import java.util.List;

public class V1beta1PodItem {
    private V1beta1PodItemsMetadata metadata;
    private String timestamp;
    private String window;
    private List<V1beta1Container> containers = new ArrayList<>();

    public V1beta1PodItemsMetadata getMetadata() {
        return metadata;
    }

    public void setMetadata(V1beta1PodItemsMetadata metadata) {
        this.metadata = metadata;
    }

    public String getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(String timestamp) {
        this.timestamp = timestamp;
    }

    public String getWindow() {
        return window;
    }

    public void setWindow(String window) {
        this.window = window;
    }

    public List<V1beta1Container> getContainers() {
        return containers;
    }

    public void setContainers(List<V1beta1Container> containers) {
        this.containers = containers;
    }
}
