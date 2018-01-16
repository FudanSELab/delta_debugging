package apiserver.bean;

import java.util.List;

public class QueryDeploymentsListResponse {
    private String apiVersion = null;

    private String kind = null;

    private List<SingleDeploymentInfo> items;

    public QueryDeploymentsListResponse(){

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

    public List<SingleDeploymentInfo> getItems() {
        return items;
    }

    public void setItems(List<SingleDeploymentInfo> items) {
        this.items = items;
    }
}
