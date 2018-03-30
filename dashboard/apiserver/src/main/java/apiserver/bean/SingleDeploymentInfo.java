package apiserver.bean;

public class SingleDeploymentInfo {
    private V1ObjectMeta metadata;
    private AppsV1beta1DeploymentSpec spec = null;
    private SingleDeploymentStatus status;

    public SingleDeploymentInfo(){

    }

    public V1ObjectMeta getMetadata() {
        return metadata;
    }

    public void setMetadata(V1ObjectMeta metadata) {
        this.metadata = metadata;
    }

    public AppsV1beta1DeploymentSpec getSpec() {
        return spec;
    }

    public void setSpec(AppsV1beta1DeploymentSpec spec) {
        this.spec = spec;
    }

    public SingleDeploymentStatus getStatus() {
        return status;
    }

    public void setStatus(SingleDeploymentStatus status) {
        this.status = status;
    }
}
