package apiserver.bean;

public class AppsV1beta1DeploymentStatus {
    private String replicas;
    private Selector selector;
    private String targetSelector;

    public AppsV1beta1DeploymentStatus(){

    }

    public String getReplicas() {
        return replicas;
    }

    public void setReplicas(String replicas) {
        this.replicas = replicas;
    }

    public Selector getSelector() {
        return selector;
    }

    public void setSelector(Selector selector) {
        this.selector = selector;
    }

    public String getTargetSelector() {
        return targetSelector;
    }

    public void setTargetSelector(String targetSelector) {
        this.targetSelector = targetSelector;
    }
}
