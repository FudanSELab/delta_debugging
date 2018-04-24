package deltabackend.domain.bean;

public class ServiceWithReplicas {
    private String serviceName;
    private int numOfReplicas;

    public ServiceWithReplicas(){

    }

    public String getServiceName() {
        return serviceName;
    }

    public void setServiceName(String serviceName) {
        this.serviceName = serviceName;
    }

    public int getNumOfReplicas() {
        return numOfReplicas;
    }

    public void setNumOfReplicas(int numOfReplicas) {
        this.numOfReplicas = numOfReplicas;
    }

    public String toString(){
        return this.serviceName + ": " + this.numOfReplicas;
    }
}
