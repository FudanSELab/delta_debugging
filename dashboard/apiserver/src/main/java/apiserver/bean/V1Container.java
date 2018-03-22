package apiserver.bean;

import java.util.List;

public class V1Container {
    private String name = null;
    private String image = null;
    private String imagePullPolicy = null;
    private List<V1ContainerPort> ports = null;

    public V1Container(){

    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getImage() {
        return image;
    }

    public void setImage(String image) {
        this.image = image;
    }

    public String getImagePullPolicy() {
        return imagePullPolicy;
    }

    public void setImagePullPolicy(String imagePullPolicy) {
        this.imagePullPolicy = imagePullPolicy;
    }

    public List<V1ContainerPort> getPorts() {
        return ports;
    }

    public void setPorts(List<V1ContainerPort> ports) {
        this.ports = ports;
    }
}
