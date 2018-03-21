package apiserver.bean;

public class V1ContainerStateRunning {
    private String startedAt = null;

    public V1ContainerStateRunning(){

    }

    public String getStartedAt() {
        return startedAt;
    }

    public void setStartedAt(String startedAt) {
        this.startedAt = startedAt;
    }
}
