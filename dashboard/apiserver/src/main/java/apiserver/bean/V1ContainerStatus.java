package apiserver.bean;

public class V1ContainerStatus {
    private String name = null;
    private Boolean ready = null;
    private Integer restartCount = null;
    private String image = null;
    private String imageID = null;
    private String containerID = null;
    private V1ContainerState state = null;
    private V1ContainerState lastState = null;

    public V1ContainerStatus(){

    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public Boolean getReady() {
        return ready;
    }

    public void setReady(Boolean ready) {
        this.ready = ready;
    }

    public Integer getRestartCount() {
        return restartCount;
    }

    public void setRestartCount(Integer restartCount) {
        this.restartCount = restartCount;
    }

    public String getImage() {
        return image;
    }

    public void setImage(String image) {
        this.image = image;
    }

    public String getImageID() {
        return imageID;
    }

    public void setImageID(String imageID) {
        this.imageID = imageID;
    }

    public String getContainerID() {
        return containerID;
    }

    public void setContainerID(String containerID) {
        this.containerID = containerID;
    }

    public V1ContainerState getState() {
        return state;
    }

    public void setState(V1ContainerState state) {
        this.state = state;
    }

    public V1ContainerState getLastState() {
        return lastState;
    }

    public void setLastState(V1ContainerState lastState) {
        this.lastState = lastState;
    }
}
