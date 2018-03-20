package apiserver.request;

public class GetSinglePodLogRequest {
    private String podName;

    public GetSinglePodLogRequest(){

    }

    public String getPodName() {
        return podName;
    }

    public void setPodName(String podName) {
        this.podName = podName;
    }
}
