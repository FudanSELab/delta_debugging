package deltabackend.domain.bean;

public class SingleDeltaCMResourceRequest {
    private String serviceName;
    private String type;
    private String key;
    private String value;

    public SingleDeltaCMResourceRequest(){

    }

    public String getServiceName() {
        return serviceName;
    }

    public void setServiceName(String serviceName) {
        this.serviceName = serviceName;
    }

    public String getType() {
        return type;
    }

    public void setType(String type) {
        this.type = type;
    }

    public String getKey() {
        return key;
    }

    public void setKey(String key) {
        this.key = key;
    }

    public String getValue() {
        return value;
    }

    public void setValue(String value) {
        this.value = value;
    }

    public String toString(){
        return this.serviceName + ": " + this.type + ": " + this.key + ": " + this.value;
    }
}
