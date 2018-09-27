package apiserver.bean.Metrics.NodeMetrics;

public class V1beta1NodeItemsUsage {
    private String cpu;
    private String memory;

    public String getCpu() {
        return cpu;
    }

    public void setCpu(String cpu) {
        this.cpu = cpu;
    }

    public String getMemory() {
        return memory;
    }

    public void setMemory(String memory) {
        this.memory = memory;
    }
}
