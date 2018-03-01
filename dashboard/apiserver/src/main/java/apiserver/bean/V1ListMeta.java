package apiserver.bean;

public class V1ListMeta {
    private String resourceVersion = null;
    private String selfLink = null;

    public V1ListMeta(){

    }

    public String getResourceVersion() {
        return resourceVersion;
    }

    public void setResourceVersion(String resourceVersion) {
        this.resourceVersion = resourceVersion;
    }

    public String getSelfLink() {
        return selfLink;
    }

    public void setSelfLink(String selfLink) {
        this.selfLink = selfLink;
    }
}
