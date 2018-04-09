package apiserver.bean;

import java.util.ArrayList;
import java.util.List;

public class V1Endpoints {
    private String apiVersion = null;
    private String kind = null;
    private V1ObjectMeta metadata = null;
    private List<V1EndpointSubset> subsets = new ArrayList<V1EndpointSubset>();

    public V1Endpoints(){

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

    public V1ObjectMeta getMetadata() {
        return metadata;
    }

    public void setMetadata(V1ObjectMeta metadata) {
        this.metadata = metadata;
    }

    public List<V1EndpointSubset> getSubsets() {
        return subsets;
    }

    public void setSubsets(List<V1EndpointSubset> subsets) {
        this.subsets = subsets;
    }
}
