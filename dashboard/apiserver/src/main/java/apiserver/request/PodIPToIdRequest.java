package apiserver.request;

import java.util.List;

public class PodIPToIdRequest {
    private String cluster;
    private List<String> ips;

    public String getCluster() {
        return cluster;
    }

    public void setCluster(String cluster) {
        this.cluster = cluster;
    }

    public List<String> getIps() {
        return ips;
    }

    public void setIps(List<String> ips) {
        this.ips = ips;
    }
}
