package apiserver.bean;

import java.util.List;
import java.util.Map;

public class V1NodeStatus {
    private List<V1NodeAddress> addresses = null;
    private Map<String, String> capacity = null;
    private Map<String, String> allocatable = null;
    private V1NodeSystemInfo nodeInfo = null;
    private List<V1NodeCondition> conditions = null;

    public V1NodeStatus(){

    }

    public List<V1NodeAddress> getAddresses() {
        return addresses;
    }

    public void setAddresses(List<V1NodeAddress> addresses) {
        this.addresses = addresses;
    }

    public Map<String, String> getCapacity() {
        return capacity;
    }

    public void setCapacity(Map<String, String> capacity) {
        this.capacity = capacity;
    }

    public Map<String, String> getAllocatable() {
        return allocatable;
    }

    public void setAllocatable(Map<String, String> allocatable) {
        this.allocatable = allocatable;
    }

    public V1NodeSystemInfo getNodeInfo() {
        return nodeInfo;
    }

    public void setNodeInfo(V1NodeSystemInfo nodeInfo) {
        this.nodeInfo = nodeInfo;
    }

    public List<V1NodeCondition> getConditions() {
        return conditions;
    }

    public void setConditions(List<V1NodeCondition> conditions) {
        this.conditions = conditions;
    }
}
