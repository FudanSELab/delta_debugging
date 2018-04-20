package deltabackend.domain.api.request;

public class GetSinglePodLogRequest {
    private String clusterName;
    private String podName;

    public GetSinglePodLogRequest(){

    }

    public String getClusterName() {
        return clusterName;
    }

    public void setClusterName(String clusterName) {
        this.clusterName = clusterName;
    }

    public String getPodName() {
        return podName;
    }

    public void setPodName(String podName) {
        this.podName = podName;
    }
}
