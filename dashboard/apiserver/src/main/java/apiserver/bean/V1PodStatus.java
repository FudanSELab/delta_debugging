package apiserver.bean;

import java.util.List;

public class V1PodStatus {
    private String phase = null;
    private String hostIP = null;
    private String podIP = null;
    private String startTime = null;
    private List<V1ContainerStatus> containerStatuses = null;

    public V1PodStatus(){

    }

    public String getPhase() {
        return phase;
    }

    public void setPhase(String phase) {
        this.phase = phase;
    }

    public String getHostIP() {
        return hostIP;
    }

    public void setHostIP(String hostIP) {
        this.hostIP = hostIP;
    }

    public String getPodIP() {
        return podIP;
    }

    public void setPodIP(String podIP) {
        this.podIP = podIP;
    }

    public String getStartTime() {
        return startTime;
    }

    public void setStartTime(String startTime) {
        this.startTime = startTime;
    }

    public List<V1ContainerStatus> getContainerStatuses() {
        return containerStatuses;
    }

    public void setContainerStatuses(List<V1ContainerStatus> containerStatuses) {
        this.containerStatuses = containerStatuses;
    }
}
