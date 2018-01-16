package apiserver.bean;

public class AppsV1beta1DeploymentSpec {
    private String replicas;

    public AppsV1beta1DeploymentSpec(){

    }

    public String getReplicas() {
        return replicas;
    }

    public void setReplicas(String replicas) {
        this.replicas = replicas;
    }
}
