package apiserver.bean.Metrics.NodeMetrics;

import apiserver.bean.Metrics.Common.V1beta1ItemsUsage;

public class V1beta1NodeItem {
    private V1beta1NodeItemsMetadata metadata;
    private String timestamp;
    private String window;
    private V1beta1ItemsUsage usage;

    public V1beta1NodeItemsMetadata getMetadata() {
        return metadata;
    }

    public void setMetadata(V1beta1NodeItemsMetadata metadata) {
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

    public V1beta1ItemsUsage getUsage() {
        return usage;
    }

    public void setUsage(V1beta1ItemsUsage usage) {
        this.usage = usage;
    }
}
