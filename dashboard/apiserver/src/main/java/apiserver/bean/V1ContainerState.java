package apiserver.bean;

public class V1ContainerState {
    private V1ContainerStateRunning running = null;
    private V1ContainerStateTerminated terminated = null;
    private V1ContainerStateWaiting waiting = null;

    public V1ContainerState(){

    }

    public V1ContainerStateRunning getRunning() {
        return running;
    }

    public void setRunning(V1ContainerStateRunning running) {
        this.running = running;
    }

    public V1ContainerStateTerminated getTerminated() {
        return terminated;
    }

    public void setTerminated(V1ContainerStateTerminated terminated) {
        this.terminated = terminated;
    }

    public V1ContainerStateWaiting getWaiting() {
        return waiting;
    }

    public void setWaiting(V1ContainerStateWaiting waiting) {
        this.waiting = waiting;
    }
}
