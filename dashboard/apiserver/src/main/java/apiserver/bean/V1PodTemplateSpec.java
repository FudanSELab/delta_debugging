package apiserver.bean;

public class V1PodTemplateSpec {
    private V1ObjectMeta metadata = null;
    private V1PodSpec spec = null;

    public V1PodTemplateSpec(){

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
}
