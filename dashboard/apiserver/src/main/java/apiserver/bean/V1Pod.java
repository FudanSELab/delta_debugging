package apiserver.bean;

public class V1Pod {
    private V1ObjectMeta metadata = null;
    private V1PodSpec spec = null;
    private V1PodStatus status = null;

    public V1Pod(){

    }

    public V1ObjectMeta getMetadata() {
        return metadata;
    }

    public void setMetadata(V1ObjectMeta metadata) {
        this.metadata = metadata;
    }

    public V1PodSpec getSpec() {
        return spec;
    }

    public void setSpec(V1PodSpec spec) {
        this.spec = spec;
    }

    public V1PodStatus getStatus() {
        return status;
    }

    public void setStatus(V1PodStatus status) {
        this.status = status;
    }
}
