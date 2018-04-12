package apiserver.response;

import apiserver.util.Cluster;

import java.util.List;

public class GetClustersResponse {
    private boolean status;
    private String message;
    private List<Cluster> clusters;

    public GetClustersResponse(){

    }

    public boolean isStatus() {
        return status;
    }

    public void setStatus(boolean status) {
        this.status = status;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }

    public List<Cluster> getClusters() {
        return clusters;
    }

    public void setClusters(List<Cluster> clusters) {
        this.clusters = clusters;
    }
}
