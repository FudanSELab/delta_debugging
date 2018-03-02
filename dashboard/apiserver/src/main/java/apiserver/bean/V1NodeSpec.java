package apiserver.bean;

import java.util.List;

public class V1NodeSpec {
    private String externalID = null;
    private String podCIDR = null;
    private List<V1Taint> taints = null;

    public V1NodeSpec(){

    }

    public String getExternalID() {
        return externalID;
    }

    public void setExternalID(String externalID) {
        this.externalID = externalID;
    }

    public String getPodCIDR() {
        return podCIDR;
    }

    public void setPodCIDR(String podCIDR) {
        this.podCIDR = podCIDR;
    }

    public List<V1Taint> getTaints() {
        return taints;
    }

    public void setTaints(List<V1Taint> taints) {
        this.taints = taints;
    }
}
