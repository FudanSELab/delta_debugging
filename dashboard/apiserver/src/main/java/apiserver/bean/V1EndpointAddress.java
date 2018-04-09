package apiserver.bean;

public class V1EndpointAddress {
    private String hostname = null;
    private String ip = null;
    private String nodeName = null;
    private V1ObjectReference targetRef = null;

    public V1EndpointAddress(){

    }

    public String getHostname() {
        return hostname;
    }

    public void setHostname(String hostname) {
        this.hostname = hostname;
    }

    public String getIp() {
        return ip;
    }

    public void setIp(String ip) {
        this.ip = ip;
    }

    public String getNodeName() {
        return nodeName;
    }

    public void setNodeName(String nodeName) {
        this.nodeName = nodeName;
    }

    public V1ObjectReference getTargetRef() {
        return targetRef;
    }

    public void setTargetRef(V1ObjectReference targetRef) {
        this.targetRef = targetRef;
    }
}
