package apiserver.bean;

import java.util.List;

public class V1EndpointSubset {
    private List<V1EndpointAddress> addresses = null;
    private List<V1EndpointAddress> notReadyAddresses = null;
    private List<V1EndpointPort> ports = null;

    public V1EndpointSubset(){

    }

    public List<V1EndpointAddress> getAddresses() {
        return addresses;
    }

    public void setAddresses(List<V1EndpointAddress> addresses) {
        this.addresses = addresses;
    }

    public List<V1EndpointAddress> getNotReadyAddresses() {
        return notReadyAddresses;
    }

    public void setNotReadyAddresses(List<V1EndpointAddress> notReadyAddresses) {
        this.notReadyAddresses = notReadyAddresses;
    }

    public List<V1EndpointPort> getPorts() {
        return ports;
    }

    public void setPorts(List<V1EndpointPort> ports) {
        this.ports = ports;
    }
}
