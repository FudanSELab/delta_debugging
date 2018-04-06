package deltabackend.domain.instanceDelta;

import java.util.List;

public class SimpleInstanceRequest {

    String id;

    List<String> services;

    int instanceNum;


    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public List<String> getServices() {
        return services;
    }

    public void setServices(List<String> services) {
        this.services = services;
    }

    public int getInstanceNum() {
        return instanceNum;
    }

    public void setInstanceNum(int instanceNum) {
        this.instanceNum = instanceNum;
    }

}
