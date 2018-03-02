package apiserver.bean;

public class V1Node {
    private String apiVersion = null;
    private String kind = null;
    private V1NodeMeta metadata = null;
    private V1NodeSpec spec = null;

    public V1Node(){

    }

    public String getApiVersion() {
        return apiVersion;
    }

    public void setApiVersion(String apiVersion) {
        this.apiVersion = apiVersion;
    }

    public String getKind() {
        return kind;
    }

    public void setKind(String kind) {
        this.kind = kind;
    }

    public V1NodeMeta getMetadata() {
        return metadata;
    }

    public void setMetadata(V1NodeMeta metadata) {
        this.metadata = metadata;
    }

    public V1NodeSpec getSpec() {
        return spec;
    }

    public void setSpec(V1NodeSpec spec) {
        this.spec = spec;
    }

}
