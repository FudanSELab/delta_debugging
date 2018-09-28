package apiserver.bean.Metrics.PodMetrics;

import apiserver.bean.Metrics.Common.V1beta1ItemsUsage;

public class V1beta1Container {
    private String name;
    private V1beta1ItemsUsage usage;

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public V1beta1ItemsUsage getUsage() {
        return usage;
    }

    public void setUsage(V1beta1ItemsUsage usage) {
        this.usage = usage;
    }
}
