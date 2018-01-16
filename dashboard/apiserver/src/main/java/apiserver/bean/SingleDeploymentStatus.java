package apiserver.bean;

public class SingleDeploymentStatus {
    private int observedGeneration;
    private int replicas;
    private int updatedReplicas;
    private int readyReplicas;
    private int availableReplicas;

    public SingleDeploymentStatus(){

    }

    public int getObservedGeneration() {
        return observedGeneration;
    }

    public void setObservedGeneration(int observedGeneration) {
        this.observedGeneration = observedGeneration;
    }

    public int getReplicas() {
        return replicas;
    }

    public void setReplicas(int replicas) {
        this.replicas = replicas;
    }

    public int getUpdatedReplicas() {
        return updatedReplicas;
    }

    public void setUpdatedReplicas(int updatedReplicas) {
        this.updatedReplicas = updatedReplicas;
    }

    public int getReadyReplicas() {
        return readyReplicas;
    }

    public void setReadyReplicas(int readyReplicas) {
        this.readyReplicas = readyReplicas;
    }

    public int getAvailableReplicas() {
        return availableReplicas;
    }

    public void setAvailableReplicas(int availableReplicas) {
        this.availableReplicas = availableReplicas;
    }
}
