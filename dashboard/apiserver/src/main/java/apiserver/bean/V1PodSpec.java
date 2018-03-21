package apiserver.bean;

import java.util.ArrayList;
import java.util.List;

public class V1PodSpec {
    private String nodeName = null;
    private String schedulerName = null;
    private List<V1Container> containers = new ArrayList<V1Container>();

    public V1PodSpec(){

    }

    public String getNodeName() {
        return nodeName;
    }

    public void setNodeName(String nodeName) {
        this.nodeName = nodeName;
    }

    public String getSchedulerName() {
        return schedulerName;
    }

    public void setSchedulerName(String schedulerName) {
        this.schedulerName = schedulerName;
    }

    public List<V1Container> getContainers() {
        return containers;
    }

    public void setContainers(List<V1Container> containers) {
        this.containers = containers;
    }
}
