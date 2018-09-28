package apiserver.bean.Metrics.PodMetrics;

import apiserver.bean.Metrics.NodeMetrics.V1beta1NodeItemsUsage;

public class V1beta1Container {
    private String name;
    private V1beta1NodeItemsUsage usage;

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public V1beta1NodeItemsUsage getUsage() {
        return usage;
    }

    public void setUsage(V1beta1NodeItemsUsage usage) {
        this.usage = usage;
    }
}
