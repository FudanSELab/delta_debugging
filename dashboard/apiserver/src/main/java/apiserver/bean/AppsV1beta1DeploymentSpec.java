package apiserver.bean;

public class AppsV1beta1DeploymentSpec {
    private String replicas;
    private V1PodTemplateSpec template = null;

    public AppsV1beta1DeploymentSpec(){

    }

    public String getReplicas() {
        return replicas;
    }

    public void setReplicas(String replicas) {
        this.replicas = replicas;
    }

    public V1PodTemplateSpec getTemplate() {
        return template;
    }

    public void setTemplate(V1PodTemplateSpec template) {
        this.template = template;
    }
}
