package apiserver.bean;

import java.util.ArrayList;
import java.util.List;

public class V1EndpointsList {
    private String apiVersion = null;
    private List<V1Endpoints> items = new ArrayList<V1Endpoints>();
    private String kind = null;
    private V1ListMeta metadata = null;

    public V1EndpointsList(){

    }

    public String getApiVersion() {
        return apiVersion;
    }

    public void setApiVersion(String apiVersion) {
        this.apiVersion = apiVersion;
    }

    public List<V1Endpoints> getItems() {
        return items;
    }

    public void setItems(List<V1Endpoints> items) {
        this.items = items;
    }

    public String getKind() {
        return kind;
    }

    public void setKind(String kind) {
        this.kind = kind;
    }

    public V1ListMeta getMetadata() {
        return metadata;
    }

    public void setMetadata(V1ListMeta metadata) {
        this.metadata = metadata;
    }
}
